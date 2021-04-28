package main

import (
	"errors"
	"lib"
	"net/url"
	"strconv"
	"time"

	"github.com/thin-peak/httpservice"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
	"github.com/thin-peak/logger"
)

type User struct {
	Id       string `bson:"_id"`
	Mail     string `bson:"mail"`
	Name     string `bson:"name"`
	Surname  string `bson:"surname"`
	Otch     string `bson:"otch"`
	Position string `bson:"position"`
	Kaf      string `bson:"kaf"`
	Fac      string `bson:"fac"`
}

type RegisterWithFormHandler struct {
	mgoSession      *mgo.Session
	mgoColl         *mgo.Collection
	trntlConn       *tarantool.Connection
	trntlTable      string
	trntlTableCodes string
	configurator    *httpservice.Configurator
	logWriters      []logger.LogWriter
}

func (handler *RegisterWithFormHandler) Close() error {
	handler.mgoSession.Close()
	return handler.trntlConn.Close()
}

func (flags *flags) NewHandler(configurator *httpservice.Configurator) (*RegisterWithFormHandler, error) {

	logWriters, err := lib.LogsInit(configurator)
	if err != nil {
		return nil, err
	}

	trntlConnection, err := tarantool.Connect(flags.trntlAddr, tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		return nil, err
	}

	mgoSession, err := mgo.Dial("127.0.0.1")
	if err != nil {
		return nil, err
	}

	return &RegisterWithFormHandler{mgoSession: mgoSession, mgoColl: mgoSession.DB("main").C(flags.mgoCollName), trntlConn: trntlConnection, trntlTable: flags.trntlTable, configurator: configurator, logWriters: logWriters}, nil
}

func (handler *RegisterWithFormHandler) Handle(r *suckhttp.Request) (w *suckhttp.Response, err error) {

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}

	userRegCodeStr := formValues.Get("code")
	if userRegCodeStr == "" {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}
	foo, err := strconv.ParseInt(userRegCodeStr, 10, 32)
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return w, err
	}
	userRegCodeInt := int32(foo)

	var trntlRes []interface{}
	err = handler.trntlConn.SelectTyped(handler.trntlTableCodes, "primary", 0, 1, tarantool.IterEq, []interface{}{userRegCodeInt}, &trntlRes)
	if err != nil {
		return nil, err
	}

	if len(trntlRes) == 0 {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}

	userFPassword := formValues.Get("password1")
	userSPassword := formValues.Get("password2") // чтоб наверняка??
	if len(userFPassword) < 8 || len(userFPassword) > 40 {
		w.SetStatusCode(418, "") // TODO
		return w, errors.New("unsuitable length of password")
	}
	if userFPassword != userSPassword {
		w.SetStatusCode(418, "") // TODO
		return nil, errors.New("passwords dont match")
	}
	userF := formValues.Get("f")
	userI := formValues.Get("i")
	userO := formValues.Get("o")

	if len(userF) < 2 || len(userI) < 5 || len(userO) < 5 || len(userF) > 30 || len(userI) > 30 || len(userO) > 30 {
		w.SetStatusCode(418, "") // TODO
		return nil, errors.New("too short f/i/o")
	}

	userPosition := formValues.Get("position") // TODO: users position
	userKaf := formValues.Get("kaf")           // TODO: kafedra
	userFac := formValues.Get("fac")           // TODO: faculty

	userMail := formValues.Get("mail")
	if !lib.IsEmailValid(userMail) {
		w.SetStatusCode(418, "") // TODO
		return w, errors.New("email isnt valid")
	}
	userMailHash, err := lib.GetMD5(userMail)
	if err != nil {
		return nil, err
	}
	userPassHash, err := lib.GetMD5(userFPassword)
	if err != nil {
		return nil, err
	}

	_, err = handler.trntlConn.Insert(handler.trntlTable, []interface{}{userMailHash, userPassHash})
	if err != nil {
		if errors.Is(err, tarantool.Error{Msg: suckutils.ConcatThree("Duplicate key exists in unique index 'primary' in space '", handler.trntlTable, "'"), Code: tarantool.ErrTupleFound}) {
			w.SetStatusCode(418, "") // TODO
			return
		}
		return nil, err
	}
	insertData := &User{Id: userMailHash, Mail: userMail, Surname: userF, Name: userI, Otch: userO, Position: userPosition, Kaf: userKaf, Fac: userFac}
	err = handler.mgoColl.Insert(insertData)
	if err != nil {
		_, errr := handler.trntlConn.Delete(handler.trntlTable, "primary", []interface{}{userMailHash})
		if errr != nil {
			return nil, errr
		}
		return nil, err
	}

	// TODO: письмо на мыло

	w.SetStatusCode(200, "OK")
	return
}
