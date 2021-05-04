package main

import (
	"net/url"
	"thin-peak/logs/logger"
	"time"

	"github.com/big-larry/suckhttp"
	"github.com/big-larry/suckutils"
	"github.com/dgrijalva/jwt-go"
)

type claims struct {
	Login string
	Salt  string
	jwt.StandardClaims
}

type CookieGenerator struct {
}

func NewCookieGenerator() (*CookieGenerator, error) {

	return &CookieGenerator{}, nil
}

var jwtKey = []byte{79, 76, 69, 71}

func (conf *CookieGenerator) Handle(r *suckhttp.Request, l *logger.Logger) (w *suckhttp.Response, err error) {

	// Это если POST-запрос и Content-Type: application/x-www-form-urlencoded
	// Можно на всякий случай проверочку сделать или еще лучше рассмотреть и реализовать варианты обращений
	formValues, err := url.ParseQuery(string(r.Body))
	if err != nil {
		return nil, err
	}
	userLoginHash := formValues.Get("login")
	if len(userLoginHash) != 32 {
		w.SetStatusCode(400, "Bad Request")
	}

	// Salt и SignedString оинаковые, так и должно быть?
	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{Login: userLoginHash, Salt: string([]byte{79, 76, 69, 71})}).SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	expires := time.Now().Add(20 * time.Hour).String()

	w.SetHeader("Set-Cookie", suckutils.ConcatFour("koki=", jwtToken, ";Expires=", expires))
	w.SetStatusCode(200, "OK")

	return w, nil
}
