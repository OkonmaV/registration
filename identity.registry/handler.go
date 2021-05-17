package main

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
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

type RegisterWithForm struct {
	mgoSession      *mgo.Session
	mgoColl         *mgo.Collection
	trntlConn       *tarantool.Connection
	trntlTable      string
	trntlCodesTable string
	tokenGenerator  *httpservice.InnerService
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// структура таблицы с кодами для регистрации
type codesTuple struct {
	code    int32
	expires int64
}

func NewRegisterWithForm(trntlAddr string, trntlTable string, trntlCodesTable string, mgoAddr string, mgoColl string, tokenGenerator *httpservice.InnerService) (*RegisterWithForm, error) {

	trntlConnection, err := tarantool.Connect(trntlAddr, tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		return nil, err
	}

	mgoSession, err := mgo.Dial(mgoAddr)
	if err != nil {
		return nil, err
	}

	return &RegisterWithForm{mgoSession: mgoSession, mgoColl: mgoSession.DB("main").C(mgoColl), trntlConn: trntlConnection, trntlTable: trntlTable, tokenGenerator: tokenGenerator}, nil
}

func (c *RegisterWithForm) Close() error {
	c.mgoSession.Close()
	return c.trntlConn.Close()
}

func (conf *RegisterWithForm) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	if !strings.Contains(r.GetHeader(suckhttp.Content_Type), "application/x-www-form-urlencoded") {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}
	// code check
	userRegCodeStr := formValues.Get("code")
	if userRegCodeStr == "" {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}
	foo, err := strconv.ParseInt(userRegCodeStr, 10, 32)
	if err != nil {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}
	userRegCodeInt := int32(foo)

	var trntlRes []interface{}
	err = conf.trntlConn.SelectTyped(conf.trntlCodesTable, "primary", 0, 1, tarantool.IterEq, []interface{}{userRegCodeInt}, &trntlRes)
	if err != nil {
		return nil, err
	}

	// TODO: (для проверки  expires) - я хз как интерфейс перегнать trntlRes в codesTuple, ибо если результат пихать в конкретную структуру,
	// то в данном случае она всегда будет [0 0].
	// И я хз как получить доступ к конкретному полю нынешнего интерфейса, если не десериализовывать

	if len(trntlRes) == 0 {
		return suckhttp.NewResponse(403, "Forbidden"), nil
	}

	// user info get & check
	userFPassword := formValues.Get("password")

	userF := formValues.Get("f")
	userI := formValues.Get("i")
	userO := formValues.Get("o")

	if len(userF) < 2 || len(userI) < 5 || len(userO) < 5 || len(userF) > 30 || len(userI) > 30 || len(userO) > 30 {
		return suckhttp.NewResponse(400, "Bad Request"), nil // TODO: bad request ли?
	}

	userPosition := formValues.Get("position") // TODO: users position
	userKaf := formValues.Get("kaf")           // TODO: kafedra
	userFac := formValues.Get("fac")           // TODO: faculty

	userMail := formValues.Get("mail")
	if !isEmailValid(userMail) {
		return suckhttp.NewResponse(400, "Bad Request"), nil // TODO: bad request ли?
	}

	userMailHash, err := getMD5(userMail)
	if err != nil {
		return nil, err
	}
	userPassHash, err := getMD5(userFPassword)
	if err != nil {
		return nil, err
	}

	// tarantool insert
	_, err = conf.trntlConn.Insert(conf.trntlTable, []interface{}{userMailHash, userPassHash})
	if err != nil {
		if tarErr, ok := err.(*tarantool.Error); ok && tarErr.Code == tarantool.ErrTupleFound {
			return suckhttp.NewResponse(400, "Bad Request"), nil // TODO: bad request ли?
		}
		return nil, err
	}
	// mongo insert
	insertData := &User{Id: userMailHash, Mail: userMail, Surname: userF, Name: userI, Otch: userO, Position: userPosition, Kaf: userKaf, Fac: userFac}

	err = conf.mgoColl.Insert(insertData)
	if err != nil {
		_, errr := conf.trntlConn.Delete(conf.trntlTable, "primary", []interface{}{userMailHash})
		if errr != nil {
			l.Error("Mongo insert", err)
			return nil, errr
		}
		return nil, err
	}

	// make user's cookie
	tokenReq, err := conf.tokenGenerator.CreateRequestFrom(suckhttp.GET, suckutils.ConcatTwo("/?hash=", userMailHash), r)
	if err != nil {
		return nil, err
	}
	tokenResp, err := conf.tokenGenerator.Send(tokenReq)
	if err != nil {
		return nil, err
	}
	if i, _ := tokenResp.GetStatus(); i != 200 {
		return nil, nil
	}

	expires := time.Now().Add(20 * time.Hour).String()
	resp := suckhttp.NewResponse(200, "OK")
	resp.SetHeader(suckhttp.Set_Cookie, suckutils.ConcatFour("koki=", string(tokenResp.GetBody()), ";Expires=", expires))

	// TODO: письмо на мыло

	return resp, nil
}

func getMD5(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isEmailValid(email string) bool {
	if len(email) < 6 && len(email) > 40 {
		return false
	}
	if !emailRegex.MatchString(email) {
		return false
	}
	parts := strings.Split(email, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}
