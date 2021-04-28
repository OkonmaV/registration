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

type flags struct {
	trntlAddr  string
	trntlTable string
}

func main() {

	servAddr := flag.String("listen", "", "Service listen address (unix/udp/tcp)")
	flgs := &flags{}
	flag.StringVar(&flgs.trntlAddr, "trntl", "127.0.0.1:3301", "Tarantool listener address (unix/tcp)")
	flag.StringVar(&flgs.trntlTable, "trntl", "regcodes", "Tarantool table name (string)")
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

	conf, err := httpservice.NewConfigurator(ctx, lib.ServiceNameCodesGenerator, *servAddr, httpservice.ServiceName(lib.ServiceNameCodesGenerator))
	if err != nil {
		logger.Error("Configurator connect", err)
		return
	}

	handler, err := flgs.NewHandler(conf)
	if err != nil {
		logger.Error("Init", err)
		return
	}
	defer handler.Close()
	if err := httpservice.ServeHTTPService(ctx, (*servAddr)[:strings.Index(*servAddr, ":")], (*servAddr)[strings.Index(*servAddr, ":")+1:], true, 10, handler); err != nil {
		logger.Error("Start service", err)
	}
}
