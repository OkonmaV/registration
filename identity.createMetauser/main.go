package main

import (
	"context"

	"thin-peak/httpservice"
)

type config struct {
	Configurator string
	Listen       string
	TrntlAddr    string
	TrntlTable   string
	MgoAddr      string
	MgoColl      string
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {
	return NewCreateMetauser(c.TrntlAddr, c.TrntlTable, c.MgoAddr, c.MgoColl)
}

func main() {
	httpservice.InitNewService("conf.createmetauser", false, 5, &config{})
}
