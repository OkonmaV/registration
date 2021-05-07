package main

import (
	"errors"
	"net/url"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/dgrijalva/jwt-go"
)

type claims struct {
	Login string
	jwt.StandardClaims
}

type CookieGenerator struct {
}

func NewCookieGenerator() (*CookieGenerator, error) {

	return &CookieGenerator{}, nil
}

var jwtKey = []byte{79, 76, 69, 71}

func (conf *CookieGenerator) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	w = &suckhttp.Response{}

	if r.GetHeader("Content-Type") != "application/x-www-form-urlencoded" && r.GetMethod() != suckhttp.POST {
		w.SetStatusCode(400, "Bad Request")
		err = errors.New("wrong request's method or content-type")
		//l.Warning("Request's params", "Wrong method or content-type")
		return
	}

	// Это если POST-запрос и Content-Type: application/x-www-form-urlencoded
	// Можно на всякий случай проверочку сделать или еще лучше рассмотреть и реализовать варианты обращений
	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		w.SetStatusCode(400, "Bad Request")
		return
	}
	userLoginHash := formValues.Get("loginHash")
	userLogin := formValues.Get("login")
	if len(userLoginHash) != 32 || userLogin == "" {
		w.SetStatusCode(400, "Bad Request")
		return w, nil
	}

	var jwtToken string

	if userLogin != "" || userLoginHash == "" {

		jwtToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{Login: userLogin}).SignedString(jwtKey)
		if err != nil {
			return nil, err
		}
	} else if userLogin == "" || userLoginHash != "" {

		jwtToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{Login: userLoginHash}).SignedString(jwtKey)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("all params aren't null") // ? ="вы там с ума сходите?"
	}

	expires := time.Now().Add(20 * time.Hour).String()

	w.SetHeader(suckhttp.Set_Cookie, suckutils.ConcatFour("koki=", jwtToken, ";Expires=", expires))
	w.SetStatusCode(200, "OK")

	return w, nil
}
