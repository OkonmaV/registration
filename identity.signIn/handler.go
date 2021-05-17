package main

import (
	"lib"
	"net/url"
	"strings"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
)

type Authentication struct {
	trntlConn      *tarantool.Connection
	trntlTable     string
	tokenGenerator *httpservice.InnerService
}

func (handler *Authentication) Close() error {
	return handler.trntlConn.Close()
}

func NewAuthentication(trntlAddr string, trntlTable string, tokenGenerator *httpservice.InnerService) (*Authentication, error) {

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

	return &Authentication{trntlConn: trntlConn, trntlTable: trntlTable, tokenGenerator: tokenGenerator}, nil
}

func (conf *Authentication) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	if !strings.Contains(r.GetHeader(suckhttp.Content_Type), "application/x-www-form-urlencoded") {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	formValue, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}
	login := formValue.Get("login")
	password := formValue.Get("password")
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

	tokenReq, err := conf.tokenGenerator.CreateRequestFrom(suckhttp.GET, suckutils.ConcatTwo("/?hash=", hashLogin), r)
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

	return resp, nil
}
