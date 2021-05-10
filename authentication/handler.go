package main

import (
	"lib"
	"net/url"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
)

type Authentication struct {
	trntlConn            *tarantool.Connection
	trntlTable           string
	cookieTokenGenerator *httpservice.InnerService
}

func (handler *Authentication) Close() error {
	return handler.trntlConn.Close()
}

func NewAuthentication(trntlAddr string, trntlTable string, cookieTokenGenerator *httpservice.InnerService) (*Authentication, error) {

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

	return &Authentication{trntlConn: trntlConn, trntlTable: trntlTable, cookieTokenGenerator: cookieTokenGenerator}, nil
}

func (conf *Authentication) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad request"), err
	}

	login := formValues.Get("login")
	password := formValues.Get("password")
	if login == "" || password == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	hashLogin, err := lib.GetMD5(login)
	if err != nil {
		return nil, err
	}
	hashPassword, err := lib.GetMD5(password)
	if err != nil {
		return nil, err
	}
	var trntlRes []interface{}
	if err = conf.trntlConn.SelectTyped(conf.trntlTable, "secondary", 0, 1, tarantool.IterEq, []interface{}{hashLogin, hashPassword}, &trntlRes); err != nil {
		return nil, err
	}
	if len(trntlRes) == 0 {
		return suckhttp.NewResponse(403, "Forbidden"), nil
	}

	cookieResp, err := conf.cookieTokenGenerator.Send(r)
	if err != nil {
		return nil, err
	}
	if i, _ := cookieResp.GetStatus(); i != 200 {
		return nil, nil
	}

	cookieTokenReq := *r
	cookieTokenReq.Body = []byte(hashLogin)
	cookieTokenResp, err := conf.cookieTokenGenerator.Send(&cookieTokenReq)
	if err != nil {
		return nil, err
	}
	if i, _ := cookieTokenResp.GetStatus(); i != 200 {
		return nil, nil
	}

	expires := time.Now().Add(20 * time.Hour).String()
	resp := suckhttp.NewResponse(200, "OK")
	resp.SetHeader(suckhttp.Set_Cookie, suckutils.ConcatFour("koki=", string(cookieTokenResp.GetBody()), ";Expires=", expires))

	return resp, nil
}
