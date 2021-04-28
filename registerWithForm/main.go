package main

import (
	"errors"
	"lib"
	"net/url"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

type database struct {
	mgoSession      *mgo.Session
	mgoColl         *mgo.Collection
	trntlConn       *tarantool.Connection
	trntlTable      string
	trntlTableCodes string
}

type User struct {
	Id       string `bson:"_id"`
	Mail     string `bson:"mail"`
	Name     string `bson:"name"`
	Surname  string `bson:"surname"`
	Otch     string `bson:"otch"`
	Position string `bson:"position"`
	Kaf      string `bson:"kaf"`
}

func connCloser(db database) error {
	db.mgoSession.Close()
	return db.trntlConn.Close()
}

func main() {
	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
	logger.SetupLogger(ctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

	handler, err := NewRegisterWithFormHandler()
	if err != nil {
		logger.Error("RegisterWithForm service", err)
		return
	}

	logger.Error("RegisterWithForm service", httpservice.ServeHTTPService(ctx, "tcp", ":8092", true, 10, handler))

	defer logger.Error("RegisterWithForm service", connCloser(*handler))
	defer func() {
		cancel()
		<-logger.AllLogsFlushed
	}()
}

func NewRegisterWithFormHandler() (*database, error) {
	mgoSession, err := mgo.Dial("127.0.0.1")
	if err != nil {
		return nil, err
	}

	trntlConnection, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		return nil, err
	}

	return &database{mgoColl: mgoSession.DB("main").C("users"), trntlConn: trntlConnection, trntlTable: "users", trntlTableCodes: "regcodes"}, nil
}

func (db *database) Handle(r *suckhttp.Request) (w *suckhttp.Response, err error) {

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}

	userRegCode := formValues.Get("code")
	if userRegCode == "" {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}
	var trntlRes []interface{}
	err = db.trntlConn.SelectTyped(db.trntlTableCodes, "primary", 0, 1, tarantool.IterEq, []interface{}{userRegCode}, &trntlRes)
	if err != nil {
		return nil, err
	}

	if len(trntlRes) == 0 {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}

	userFPassword := formValues.Get("password1")
	userSPassword := formValues.Get("password2")
	if len(userFPassword) < 8 {
		w.SetStatusCode(418, "") // TODO
		return w, errors.New("too short password")
	}
	if userFPassword != userSPassword { // TODO: ??
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

	_, err = db.trntlConn.Insert(db.trntlTable, []interface{}{userMailHash, userPassHash})
	if err != nil {
		if errors.Is(err, tarantool.Error{Msg: suckutils.ConcatThree("Duplicate key exists in unique index 'primary' in space '", db.trntlTable, "'"), Code: tarantool.ErrTupleFound}) {
			w.SetStatusCode(418, "") // TODO
			return
		}
		return nil, err
	}
	insertData := &User{Id: userMailHash, Mail: userMail, Surname: userF, Name: userI, Otch: userO, Position: userPosition, Kaf: userKaf}
	err = db.mgoColl.Insert(insertData)
	if err != nil {
		_, errr := db.trntlConn.Delete(db.trntlTable, "primary", []interface{}{userMailHash})
		if errr != nil {
			return nil, errr
		}
		return nil, err // да, возможно не летит ¯\_(ツ)_/¯
	}

	// TODO: письмо на мыло

	w.SetStatusCode(200, "OK")
	return

}
