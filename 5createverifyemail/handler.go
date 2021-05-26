package main

import (
	"crypto/md5"
	"encoding/hex"
	"thin-peak/logs/logger"
	"time"

	"thin-peak/httpservice"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
)

type CreateVerifyEmail struct {
	trntlConn  *tarantool.Connection
	trntlTable string
}

func (handler *CreateVerifyEmail) Close() error {
	return handler.trntlConn.Close()
}

func NewCreateVerifyEmail(trntlAddr string, trntlTable string, createVerifyEmail *httpservice.InnerService) (*CreateVerifyEmail, error) {

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
	logger.Info("Tarantool", "Connected!")
	return &CreateVerifyEmail{trntlConn: trntlConn, trntlTable: trntlTable}, nil
}

func (conf *CreateVerifyEmail) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	mail := string(r.Body)
	if mail == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	hashedMail, err := getMD5(mail)
	if err != nil {
		return nil, err
	}

	_, err = conf.trntlConn.Insert(conf.trntlTable, []interface{}{})
	return suckhttp.NewResponse(200, "OK"), nil
}

func getMD5(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
