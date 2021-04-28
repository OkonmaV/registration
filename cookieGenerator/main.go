package main

import (
	"context"
	"flag"
	"lib"
	"os"
	"strings"
	"time"

	"github.com/thin-peak/httpservice"
	"github.com/thin-peak/logger"
)

func main() {

	servAddr := flag.String("listen", "127.0.0.1:8084", "Service listen address (unix/udp/tcp)")
	flag.Parse()

	if *servAddr == "" {
		println("listen address not set")
		os.Exit(1)
	}

	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
	loggerctx, loggercancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		loggercancel()
		<-logger.AllLogsFlushed
	}()
	logger.SetupLogger(loggerctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

	conf, err := httpservice.NewConfigurator(ctx, lib.ServiceNameCookieGen, *servAddr, httpservice.ServiceName(lib.ServiceNameCookieGen))
	if err != nil {
		logger.Error("Configurator connect", err)
		return
	}

	handler, err := NewHandler(conf)
	if err != nil {
		logger.Error("Init", err)
		return
	}
	if err := httpservice.ServeHTTPService(ctx, (*servAddr)[:strings.Index(*servAddr, ":")], (*servAddr)[strings.Index(*servAddr, ":")+1:], false, 10, handler); err != nil {
		logger.Error("Start service", err)
	}
}
