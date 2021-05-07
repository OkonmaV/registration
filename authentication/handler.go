package main

import (
	"errors"
	"lib"
	"net/url"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
)

type Authentication struct {
	trntlConn  *tarantool.Connection
	trntlTable string
	connectors map[httpservice.ServiceName]*httpservice.InnerService
}

func (handler *Authentication) Close() error {
	return handler.trntlConn.Close()
}

func NewAuthentication(trntlAddr string, trntlTable string, conns map[httpservice.ServiceName]*httpservice.InnerService) (*Authentication, error) {

	trntlConn, err := tarantool.Connect(trntlAddr, tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		return nil, err
	}
	return &Authentication{trntlConn: trntlConn, trntlTable: trntlTable, connectors: conns}, nil
}

func (conf *Authentication) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}

	login := formValues.Get("login")
	password := formValues.Get("password")
	if login == "" || password == "" {
		w.SetStatusCode(400, "Bad Request")
		return
	}

	hashLogin, err := lib.GetMD5(login)
	if err != nil {
		logger.Error("Get MD5", err)
		return nil, err
	}
	hashPassword, err := lib.GetMD5(password)
	if err != nil {
		logger.Error("Get MD5", err)
		return nil, err
	}
	var trntlRes []interface{}
	err = conf.trntlConn.SelectTyped(conf.trntlTable, "secondary", 0, 1, tarantool.IterEq, []interface{}{hashLogin, hashPassword}, &trntlRes)
	if err != nil {
		return nil, err
	}
	if len(trntlRes) == 0 {
		w.SetStatusCode(403, "Forbidden")
		return
	}
	cookieServ := conf.connectors[lib.ServiceNameCookieGen]
	if cookieServ == nil {
		logger.Error("Inner Service Conn", errors.New("nil pointer instead if cookieGen connection"))
		return nil, errors.New("nil pointer instead if cookieGen (as inner service) connection ")
	}
	cookieResp, err := cookieServ.Send(r)
	if err != nil {
		return nil, err
	}
	if i, _ := cookieResp.GetStatus(); i != 200 {
		return nil, nil
	}
	return cookieResp, nil
}