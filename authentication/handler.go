package main

import (
	"lib"
	"net/url"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
)

type Authentication struct {
	trntlConn       *tarantool.Connection
	trntlTable      string
	cookieGenerator *httpservice.InnerService
}

func (handler *Authentication) Close() error {
	return handler.trntlConn.Close()
}

func NewAuthentication(trntlAddr string, trntlTable string, cookieGenerator *httpservice.InnerService) (*Authentication, error) {

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

	return &Authentication{trntlConn: trntlConn, trntlTable: trntlTable, cookieGenerator: cookieGenerator}, nil
}

func (conf *Authentication) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	w = &suckhttp.Response{}

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return
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
	if err = conf.trntlConn.SelectTyped(conf.trntlTable, "secondary", 0, 1, tarantool.IterEq, []interface{}{hashLogin, hashPassword}, &trntlRes); err != nil {
		return nil, err
	}
	if len(trntlRes) == 0 {
		w.SetStatusCode(403, "Forbidden")
		return
	}
	cookieResp, err := conf.cookieGenerator.Send(r)
	if err != nil {
		return nil, err
	}
	if i, _ := cookieResp.GetStatus(); i != 200 {
		return nil, nil
	}
	return cookieResp, nil
}
