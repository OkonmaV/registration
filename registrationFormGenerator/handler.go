package main

import (
	"net/url"
	"strings"
	"time"

	"github.com/big-larry/suckhttp"

	"thin-peak/logs/logger"

	"github.com/tarantool/go-tarantool"
)

type RegistrationFormGenerator struct {
	trntlConn  *tarantool.Connection
	trntlTable string
}

func (handler *RegistrationFormGenerator) Close() error {
	return handler.trntlConn.Close()
}

const form = `<form action="http://127.0.0.1:8094" method="POST"> 
	<input type="hidden" name="code" value="%regcode%">
	<input placeholder="name" name="name">
	<input placeholder="surname" name="surname">
	<input placeholder="otch" name="otch">
	<input placeholder="password1" type="password" name="password1">
	<input placeholder="password2" type="password" name="password2">
	<input type="submit" value="registry">
</form>
` // TODO: адрес прописать

func NewRegistrationFormGenerator(trntlAddr string, trntlTable string, trntlConn *tarantool.Connection) (*RegistrationFormGenerator, error) {

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
	return &RegistrationFormGenerator{trntlConn: trntlConn, trntlTable: trntlTable}, nil
}

func (conf *RegistrationFormGenerator) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	w = &suckhttp.Response{}

	queryValues, err := url.ParseQuery(r.Uri.RawQuery)
	if err != nil {
		return nil, err
	}

	regCode := queryValues.Get("code")
	if regCode == "" {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}
	var trntlRes []interface{}
	err = conf.trntlConn.SelectTyped(conf.trntlTable, "primary", 0, 1, tarantool.IterEq, []interface{}{regCode}, &trntlRes)
	if err != nil {
		return nil, err
	}

	if len(trntlRes) == 0 {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}

	w.SetBody([]byte(strings.ReplaceAll(form, "%regcode%", regCode)))
	return
}
