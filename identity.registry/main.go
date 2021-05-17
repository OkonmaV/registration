package main

import (
	"context"
	"lib"

	"thin-peak/httpservice"
)

type config struct {
	Configurator    string
	Listen          string
	TrntlAddr       string
	TrntlTable      string
	TrntlCodesTable string
	MgoAddr         string
	MgoColl         string
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {
	return NewRegisterWithForm(c.TrntlAddr, c.TrntlTable, c.TrntlCodesTable, c.MgoAddr, c.MgoColl, connectors[lib.ServiceNameTokenGen])
}

func main() {
	httpservice.InitNewService(lib.ServiceNameRegistry, false, 5, &config{}, lib.ServiceNameTokenGen)
}
