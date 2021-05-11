package main

import (
	"context"
	"lib"
	"thin-peak/httpservice"
)

type config struct {
	Configurator string
	Listen       string
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {
	return NewSignOut()
}

func main() {
	httpservice.InitNewService(lib.ServiceNameSignOut, false, 5, &config{})
}
