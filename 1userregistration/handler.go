package main

import (
	"net/url"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
)

type UserRegistration struct {
	trntlConn  *tarantool.Connection
	trntlTable string
}

func NewUserRegistration(trntlAddr string, trntlTable string) (*UserRegistration, error) {

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
	_, err = trntlConnection.Ping()
	if err != nil {
		return nil, err
	}
	logger.Info("Tarantool", "Connected!")

	return &UserRegistration{trntlConn: trntlConnection, trntlTable: trntlTable}, nil
}

func (c *UserRegistration) Close() error {
	return c.trntlConn.Close()
}

func (conf *UserRegistration) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	// if !strings.Contains(r.GetHeader(suckhttp.Content_Type), "application/x-www-form-urlencoded") {
	// 	return suckhttp.NewResponse(400, "Bad request"), nil
	// }
	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return suckhttp.NewResponse(400, "Bad Request"), err
	}
	// user info get & check
	hashLogin := formValues.Get("login")
	hashPassword := formValues.Get("password")
	if len(hashLogin) != 32 || hashPassword == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	// tarantool insert
	_, err = conf.trntlConn.Insert(conf.trntlTable, []interface{}{hashLogin, hashPassword})
	if err != nil {
		if tarErr, ok := err.(tarantool.Error); ok && tarErr.Code == tarantool.ErrTupleFound {
			return suckhttp.NewResponse(400, "Bad Request"), nil // TODO: bad request ли?
		}
		return nil, err
	}

	return suckhttp.NewResponse(200, "OK"), nil
}
