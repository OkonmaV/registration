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
	return NewRegistrationFormGenerator(c.TrntlAddr, c.TrntlTable)
}

func main() {
	httpservice.InitNewService(lib.ServiceNameCodesGenerator, false, 5, &config{})
}

// func main() {

// 	servAddr := flag.String("listen", "127.0.0.1:8082", "Service listen address (unix/udp/tcp)")
// 	flgs := &flags{}
// 	flag.StringVar(&flgs.trntlAddr, "trntl-address", "127.0.0.1:3301", "Tarantool listener address (unix/tcp)")
// 	flag.StringVar(&flgs.trntlTable, "trntl-table", "regcodes", "Tarantool table name (string)")
// 	flag.Parse()

// 	if *servAddr == "" {
// 		println("listen address not set")
// 		os.Exit(1)
// 	}
// 	ctx, cancel := httpservice.CreateContextWithInterruptSignal()
// 	loggerctx, loggercancel := context.WithCancel(context.Background())
// 	defer func() {
// 		cancel()
// 		loggercancel()
// 		<-logger.AllLogsFlushed
// 	}()
// 	logger.SetupLogger(loggerctx, time.Second*2, []logger.LogWriter{logger.NewConsoleLogWriter(logger.DebugLevel)})

// 	conf, err := httpservice.NewConfigurator(ctx, lib.ServiceNameFormGenerator, *servAddr, httpservice.ServiceName(lib.ServiceNameFormGenerator))

// 	if err != nil {
// 		logger.Error("Configurator connect", err)
// 		return
// 	}

// 	handler, err := flgs.NewHandler(conf)
// 	if err != nil {
// 		logger.Error("Init", err)
// 		return
// 	}

// 	defer handler.Close()
// 	if err := httpservice.ServeHTTPService(ctx, (*servAddr)[:strings.Index(*servAddr, ":")], (*servAddr)[strings.Index(*servAddr, ":")+1:], false, 10, handler); err != nil {
// 		logger.Error("Start service", err)
// 	}
// }
