package main

import (
	"context"

	"lib"

	"thin-peak/httpservice"
)

type config struct {
	Configurator string
	Listen       string
	TrntlAddr    string
	TrntlTable   string
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {
	return NewCodesGenerator(c.TrntlAddr, c.TrntlTable)
}

func main() {
	httpservice.InitNewService(lib.ServiceNameCodesGenerator, false, 5, &config{})
}
