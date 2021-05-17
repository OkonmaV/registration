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

type TokenDecoder struct {
	jwtKey []byte
}

func NewTokenDecoder(jwtKey string) (*TokenDecoder, error) {
	return &TokenDecoder{jwtKey: []byte(jwtKey)}, nil
}

func (conf *TokenDecoder) Handle(r *suckhttp.Request, l *logger.Logger) (*suckhttp.Response, error) {

	token := r.Uri.Query().Get("token")

	if token == "" {
		return suckhttp.NewResponse(400, "Bad request"), nil
	}

	decoded, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	resp := suckhttp.NewResponse(200, "OK")
	resp.SetBody([]byte(jwtToken))

	return resp, nil
}
