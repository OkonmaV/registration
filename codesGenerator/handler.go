package main

import (
	"errors"
	"lib"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

type CodesGeneratorHandler struct {
	trntlConn    *tarantool.Connection
	trntlTable   string
	configurator *httpservice.Configurator
	logWriters   []logger.LogWriter
}

func (handler *CodesGeneratorHandler) Close() error {
	return handler.trntlConn.Close()
}

func (flags *flags) NewHandler(configurator *httpservice.Configurator) (*CodesGeneratorHandler, error) {

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

	return &CodesGeneratorHandler{trntlConn: trntlConnection, trntlTable: flags.trntlTable, configurator: configurator, logWriters: logWriters}, nil
}

func (handler *CodesGeneratorHandler) Handle(r *suckhttp.Request) (w *suckhttp.Response, err error) {

	// TODO: AUTH

	w = &suckhttp.Response{}
	err = nil
	queryValues, err := url.ParseQuery(r.Uri.RawQuery)
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return
	}

	countString := queryValues.Get("count")

	countInt, err := strconv.Atoi(countString) // countString = "" вернет err
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return
	}
	// генерим коды
	var codes []int32
	for i := 0; i < countInt; i++ {
		c := rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(90000) + 10000
		codes = append(codes, c)
	}

	// закатываем
	var errStep int
	var expires time.Time = time.Now().Add(time.Hour * 72)
	var errDuplicateCodes = &tarantool.Error{Msg: suckutils.ConcatThree("Duplicate key exists in unique index 'primary' in space '", handler.trntlTable, "'"), Code: tarantool.ErrTupleFound}

	for i, c := range codes {
		_, err = handler.trntlConn.Insert(handler.trntlTable, []interface{}{c, expires})
		if err != nil {
			if errors.Is(err, *errDuplicateCodes) {

				cc := rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(90000) + 10000
				_, err = handler.trntlConn.Insert(handler.trntlTable, []interface{}{cc, expires})

				if err != nil {
					errStep = i
					break
				}

			} else {
				errStep = i
				break
			}
		}
	}

	// откатываем
	if err != nil {
		w = nil
		if errStep > 0 {
			for i := 0; i < errStep; i++ {
				_, errr := handler.trntlConn.Delete(handler.trntlTable, "primary", []interface{}{codes[i]})
				if errr != nil {
					return nil, errr
				}
			}
		}
		return
	}

	// TODO: ВОТ ТУТ ГЕНЕРИМ ДОКУМЕНТ И ОТДАЕМ
	w.SetStatusCode(200, "OK")
	return
}
