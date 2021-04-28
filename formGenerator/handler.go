package main

import (
	"lib"
	"net/url"
	"strings"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/thin-peak/httpservice"

	"github.com/tarantool/go-tarantool"
	"github.com/thin-peak/logger"
)

type FormGeneratorHandler struct {
	trntlConn    *tarantool.Connection
	trntlTable   string
	configurator *httpservice.Configurator
	logWriters   []logger.LogWriter
}

func (handler *FormGeneratorHandler) Close() error {
	return handler.trntlConn.Close()
}

const form = `<form action="http://localhost:8092" method="POST">
	<input type="hidden" name="code" value="%regcode%">
	<input placeholder="name" name="name">
	<input placeholder="surname" name="surname">
	<input placeholder="otch" name="otch">
	<input placeholder="password1" type="password" name="password1">
	<input placeholder="password2" type="password" name="password2">
	<input type="submit" value="registry">
</form>
`

func (flags *flags) NewHandler(configurator *httpservice.Configurator) (*FormGeneratorHandler, error) {

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

	return &FormGeneratorHandler{trntlConn: trntlConnection, trntlTable: flags.trntlTable, configurator: configurator, logWriters: logWriters}, nil
}

func (handler *FormGeneratorHandler) Handle(r *suckhttp.Request) (w *suckhttp.Response, err error) {

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
	err = handler.trntlConn.SelectTyped(handler.trntlTable, "primary", 0, 1, tarantool.IterEq, []interface{}{regCode}, &trntlRes)
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
