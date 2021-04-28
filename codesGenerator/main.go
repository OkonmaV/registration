package main

import (
	"net/url"
	"strconv"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
	"github.com/teris-io/shortid"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

const thisServiceName = "RegCodesGenerator service"

type database struct {
	trntlConn  *tarantool.Connection
	trntlTable string
}

func connCloser(db database) error {
	return db.trntlConn.Close()
}

func main() {
	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
	logger.SetupLogger(ctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

	handler, err := NewCodesGeneratorHandler()
	if err != nil {
		logger.Error(thisServiceName, err)
		return
	}
	logger.Error(thisServiceName, httpservice.ServeHTTPService(ctx, "tcp", ":8090", false, 10, handler))

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

	// TODO: AUTH

	w = &suckhttp.Response{}
	queryValues, err := url.ParseQuery(r.Uri.RawQuery)
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return w, err
	}

	countString := queryValues.Get("count")

	countInt, err := strconv.Atoi(countString) // countString = "" вернет err
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return w, err
	}
	// генерим коды
	var codes []string
	for i := 0; i < countInt; i++ {
		c, err := shortid.Generate()
		if err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}

	// закатываем
	var errStep int
	var errAtInsert error

	for i, c := range codes {
		_, err := db.trntlConn.Insert(db.trntlTable, []interface{}{c})
		if err != nil {
			if i == 0 {
				return nil, err
			}
			errStep = i
			errAtInsert = err
			break
		}
	}

	// откатываем
	if errStep > 0 {
		for i := 0; i < errStep; i++ {
			_, err := db.trntlConn.Delete(db.trntlTable, "primary", []interface{}{codes[i]})
			if err != nil {
				return nil, err
			}
		}
		return nil, errAtInsert
	}

	// TODO: ВОТ ТУТ ГЕНЕРИМ ДОКУМЕНТ И ОТДАЕМ
	w.SetStatusCode(200, "OK")
	return
}
