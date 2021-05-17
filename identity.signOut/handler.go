package main

import (
	"thin-peak/logs/logger"

	"github.com/big-larry/suckhttp"
)

type SignOut struct {
}

func NewSignOut() (*SignOut, error) {
	return &SignOut{}, nil
}

func (conf *SignOut) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	resp := suckhttp.NewResponse(200, "OK")
	resp.SetHeader(suckhttp.Set_Cookie, "koki=; Max-Age=-1")
	return resp, nil
}
