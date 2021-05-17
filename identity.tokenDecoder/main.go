package main

import (
	"context"
	"lib"
	"thin-peak/httpservice"
)

type config struct {
	Configurator string
	Listen       string
	JwtKey       string
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {
	return NewTokenDecoder(c.JwtKey)
}

func main() {
	httpservice.InitNewService(lib.ServiceNameTokenDecode, false, 5, &config{})
}
