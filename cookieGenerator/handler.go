package main

import (
	"lib"
	"net/url"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/dgrijalva/jwt-go"
	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

type Claims struct {
	Login string
	Salt  string
	jwt.StandardClaims
}

type CookieGeneratorHandler struct {
	configurator *httpservice.Configurator
	logWriters   []logger.LogWriter
}

func NewHandler(configurator *httpservice.Configurator) (*CookieGeneratorHandler, error) {

	logWriters, err := lib.LogsInit(configurator)
	if err != nil {
		return nil, err
	}

	return &CookieGeneratorHandler{configurator: configurator, logWriters: logWriters}, nil
}

func (handler *CookieGeneratorHandler) Handle(r *suckhttp.Request) (w *suckhttp.Response, err error) {

	jwtKey := []byte{79, 76, 69, 71}

	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}

	userLoginHash := formValues.Get("l")
	if len(userLoginHash) != 32 {
		w.SetStatusCode(400, "Bad Request")
	}

	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{Login: userLoginHash, Salt: string([]byte{79, 76, 69, 71})}).SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	expires := time.Now().Add(10 * time.Hour).String()

	w.SetHeader("Set-Cookie", suckutils.ConcatFour("koki=", jwtToken, ";Expires=", expires))
	w.SetStatusCode(200, "OK")

	return w, nil
}
