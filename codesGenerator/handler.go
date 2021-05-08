package main

import (
	"errors"
	"math/rand"
	"net/url"
	"strconv"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/tarantool/go-tarantool"
)

type CodesGenerator struct {
	trntlConn  *tarantool.Connection
	trntlTable string
}

func NewCodesGenerator(trntlAddr string, trntlTable string) (*CodesGenerator, error) {
	trntlConnection, err := tarantool.Connect(trntlAddr, tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		logger.Error("Tarantool Conn", err)
		return nil, err
	}
	logger.Info("Tarantool Conn", "Connected!")
	return &CodesGenerator{trntlConn: trntlConnection, trntlTable: trntlTable}, nil
}

func (handler *CodesGenerator) Close() error {
	return handler.trntlConn.Close()
}

func (conf *CodesGenerator) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	// TODO: AUTH

	queryValues, err := url.ParseQuery(r.Uri.RawQuery)
	if err != nil {
		resp := suckhttp.NewResponse(400, "Bad Request")
		return resp, err
	}

	countString := queryValues.Get("count")

	countInt, err := strconv.Atoi(countString) // countString = "" вернет err
	if err != nil {
		resp := suckhttp.NewResponse(400, "Bad Request")
		return resp, err
	}
	// генерим коды
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	codes := make([]int32, 0, countInt)
	for i := 0; i < countInt; i++ {
		codes[i] = rnd.Int31n(90000) + 10000
	}

	// закатываем
	var expires int64 = time.Now().Add(time.Hour * 72).Unix()
	var errDuplicateCodes = &tarantool.Error{Msg: suckutils.ConcatThree("Duplicate key exists in unique index 'primary' in space '", conf.trntlTable, "'"), Code: tarantool.ErrTupleFound}

	for countInt > 0 {
		r := rnd.Int31n(90000) + 10000
		_, err = conf.trntlConn.Insert(conf.trntlTable, []interface{}{r, expires})
		if err != nil {
			if errors.Is(err, *errDuplicateCodes) {
				continue
			} else {
				l.Error("Tarantool insert", err)
				break
			}
		}
		codes = append(codes, r)
		countInt--
	}

	// откатываем
	if err != nil {
		for _, c := range codes {
			_, errr := conf.trntlConn.Delete(conf.trntlTable, "primary", []interface{}{c})
			if err != nil {
				return nil, errr
			}
		}
		return nil, err
	}

	// TODO: ВОТ ТУТ ГЕНЕРИМ ДОКУМЕНТ И ОТДАЕМ
	// Генерируй и отдавай в зависимости от запрошенного в заголовке Accept
	resp := suckhttp.NewResponse(200, "OK")
	return resp, nil
}
