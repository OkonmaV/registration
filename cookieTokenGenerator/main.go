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
	return NewCookieTokenGenerator()
}

func main() {
	httpservice.InitNewService(lib.ServiceNameCodesGenerator, false, 5, &config{})
}

// servAddr := flag.String("listen", "127.0.0.1:8084", "Service listen address (unix/udp/tcp)")
// flag.Parse()

// if *servAddr == "" {
// 	println("listen address not set")
// 	os.Exit(1)
// }

// ctx, cancel := httpservice.CreateContextWithInterruptSignal()
// loggerctx, loggercancel := context.WithCancel(context.Background())
// defer func() {
// 	cancel()
// 	loggercancel()
// 	<-logger.AllLogsFlushed
// }()
// logger.SetupLogger(loggerctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

// conf, err := httpservice.NewConfigurator(ctx, lib.ServiceNameCookieGen, *servAddr, httpservice.ServiceName(lib.ServiceNameCookieGen))
// if err != nil {
// 	logger.Error("Configurator connect", err)
// 	return
// }

// handler, err := NewHandler(conf)
// if err != nil {
// 	logger.Error("Init", err)
// 	return
// }
// if err := httpservice.ServeHTTPService(ctx, (*servAddr)[:strings.Index(*servAddr, ":")], (*servAddr)[strings.Index(*servAddr, ":")+1:], false, 10, handler); err != nil {
// 	logger.Error("Start service", err)
// }
