package main

import (
	"thin-peak/logs/logger"

	"github.com/big-larry/suckhttp"
	"github.com/dgrijalva/jwt-go"
)

type claims struct {
	Login string
	jwt.StandardClaims
}

type CookieTokenGenerator struct {
}

func NewCookieTokenGenerator() (*CookieTokenGenerator, error) {
	return &CookieTokenGenerator{}, nil
}

var jwtKey = []byte{79, 76, 69, 71}

func (conf *CookieTokenGenerator) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	var jwtToken string

	hashLogin := string(r.Body)
	if len(hashLogin) != 32 {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{Login: hashLogin}).SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	resp := suckhttp.NewResponse(200, "OK")
	resp.SetBody([]byte(jwtToken))

	return resp, nil
}
