package main

import (
	"net/url"
	"strings"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

const thisServiceName = "RegFormGenerator service"

type database struct {
	trntlConn  *tarantool.Connection
	trntlTable string
}

func connCloser(db database) error {
	return db.trntlConn.Close()
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

func main() {
	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
	logger.SetupLogger(ctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

	handler, err := NewCodesGeneratorHandler()
	if err != nil {
		logger.Error(thisServiceName, err)
		return
	}
	logger.Error(thisServiceName, httpservice.ServeHTTPService(ctx, "tcp", ":8091", false, 10, handler))

	defer logger.Error(thisServiceName, connCloser(*handler))
	defer func() {
		cancel()
		<-logger.AllLogsFlushed
	}()
}

func NewCodesGeneratorHandler() (*database, error) {
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
	return &database{trntlConn: trntlConnection, trntlTable: "regcodes"}, nil
}

func (db *database) Handle(r *suckhttp.Request) (w *suckhttp.Response, err error) {

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
	err = db.trntlConn.SelectTyped(db.trntlTable, "primary", 0, 1, tarantool.IterEq, []interface{}{regCode}, &trntlRes)
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
