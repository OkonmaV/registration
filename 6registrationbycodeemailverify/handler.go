package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"thin-peak/httpservice"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/tarantool/go-tarantool"
)

type CreateVerifyEmail struct {
	trntlConn        *tarantool.Connection
	trntlTable       string
	emailVerify      *httpservice.InnerService
	userRegistration *httpservice.InnerService
	setUserData      *httpservice.InnerService
}

func (handler *CreateVerifyEmail) Close() error {
	return handler.trntlConn.Close()
}

func NewCreateVerifyEmail(trntlAddr string, trntlTable string, emailVerify *httpservice.InnerService, userRegistration *httpservice.InnerService, setUserData *httpservice.InnerService) (*CreateVerifyEmail, error) {

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
	return &CreateVerifyEmail{trntlConn: trntlConn, trntlTable: trntlTable, emailVerify: emailVerify, userRegistration: userRegistration, setUserData: setUserData}, nil
}

func (conf *CreateVerifyEmail) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	code := r.Uri.Query().Get("code")
	uuid := r.Uri.Query().Get("uuid")
	if code == "" || uuid == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}
	//
	var trntlRes 
	err := conf.trntlConn.SelectTyped("regcodes", "primary", 0, 1, tarantool.IterEq, []interface{}{28258}, &trntlRes)
	//
	emailVerifyReq, err := conf.emailVerify.CreateRequestFrom(suckhttp.POST, "", r)
	if err != nil {
		return nil, err
	}
	//
	emailVerifyReq.AddHeader(suckhttp.Content_Type, "application/json")
	emailVerifyInfo := make(map[string]string, 2)
	emailVerifyInfo["code"] = code
	emailVerifyInfo["uuid"] = uuid
	emailVerifyReq.Body, err = json.Marshal(emailVerifyInfo)
	if err != nil {
		return nil, err
	}
	emailVerifyResp, err := conf.emailVerify.Send(emailVerifyReq)
	if err != nil {
		return nil, err
	}

	if i, t := emailVerifyResp.GetStatus(); i != 200 {
		if i == 403 {
			return emailVerifyResp, nil
		}
		l.Debug("Responce from emailVerify", t)
		return nil, nil
	}
	//

	userRegistrationReq, err := conf.userRegistration.CreateRequestFrom(suckhttp.POST, "", r)
	if err != nil {
		return nil, err
	}

}

func getMD5(str string) (string, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(str))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
