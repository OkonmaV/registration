package main

import (
	"errors"
	"net/url"
	"time"

	"github.com/big-larry/mgo"
	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

type database struct {
	mgoSession *mgo.Session
	mgoColl    *mgo.Collection
	trntlConn  *tarantool.Connection
	trntlTable string
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

	logger.Error("RegisterWithForm service", httpservice.ServeHTTPService(ctx, "tcp", ":8090", true, 10, handler))

	defer logger.Error("RegisterWithForm service", connCloser(*handler)) // TODO: porn?
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

	trntlConn, err := tarantool.Connect("127.0.0.1:3301", tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		return nil, err
	}

	return &database{mgoColl: mgoSession.DB("main").C("users"), trntlConn: trntlConn, trntlTable: "users"}, nil
}

func (db *database) Handle(r *suckhttp.Request) (*suckhttp.Response, error) {

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}
	userFPassword := formValues.Get("password1")
	userSPassword := formValues.Get("password2")
	if len(userFPassword) < 8 {
		return nil, errors.New("too short password")
	}
	if userFPassword != userSPassword { // TODO: ??
		return nil, errors.New("passwords dont match")
	}
	userF := formValues.Get("f")
	userI := formValues.Get("i")
	userO := formValues.Get("o")

	if len(userF) < 2 || len(userI) < 5 || len(userO) < 5 || len(userF) > 30 || len(userI) > 30 || len(userO) > 30 {

	}

}
