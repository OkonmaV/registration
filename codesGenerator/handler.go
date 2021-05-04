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

func NewCodesGenerator(trntlAddr string, trntlTable string, trntlConn *tarantool.Connection) (*CodesGenerator, error) {

	trntlConnection, err := tarantool.Connect(trntlAddr, tarantool.Opts{
		// User: ,
		// Pass: ,
		Timeout:       500 * time.Millisecond,
		Reconnect:     1 * time.Second,
		MaxReconnects: 4,
	})
	if err != nil {
		return nil, err
	}

	return &CodesGenerator{trntlConn: trntlConnection, trntlTable: trntlTable}, nil
}

func (conf *CodesGenerator) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

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
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	codes := make([]int32, countInt)
	for i := 0; i < countInt; i++ {
		codes[i] = rnd.Int31n(90000) + 10000
	}

	// закатываем
	var errStep int
	var expires time.Time = time.Now().Add(time.Hour * 72)
	var errDuplicateCodes = &tarantool.Error{Msg: suckutils.ConcatThree("Duplicate key exists in unique index 'primary' in space '", conf.trntlTable, "'"), Code: tarantool.ErrTupleFound}

	// А у тарантула нету что-то вроде InsertMany ?
	for i, c := range codes {
		_, err = conf.trntlConn.Insert(conf.trntlTable, []interface{}{c, expires})
		if err != nil {
			if errors.Is(err, *errDuplicateCodes) {

				cc := rnd.Int31n(90000) + 10000
				_, err = conf.trntlConn.Insert(conf.trntlTable, []interface{}{cc, expires})

				if err != nil {
					errStep = i
					break
				}

			} else {
				logger.Warning("tarantool.Insert", err.Error())
				errStep = i
				break
			}
		}
	}

	// откатываем
	// Вот тут не понял... Что откатываем?
	if err != nil {
		w = nil
		if errStep > 0 {
			for i := 0; i < errStep; i++ {
				_, errr := conf.trntlConn.Delete(conf.trntlTable, "primary", []interface{}{codes[i]})
				if errr != nil {
					return nil, errr
				}
			}
		}
		return
	}

	// TODO: ВОТ ТУТ ГЕНЕРИМ ДОКУМЕНТ И ОТДАЕМ
	// Генерируй и отдавай в зависимости от запрошенного в заголовке Accept
	w.SetStatusCode(200, "OK")
	return
}
