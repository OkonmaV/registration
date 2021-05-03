package main

import (
	"context"
	"lib"

	"thin-peak/httpservice"

	"github.com/big-larry/mgo"
	"github.com/tarantool/go-tarantool"
)

type config struct {
	Configurator string
	Listen       string
	TrntlAddr    string
	TrntlTable   string
	TrntlConn    *tarantool.Connection
	MgoAddr      string
	MgoColl      string
	MgoConn      *mgo.Session
}

func (c *config) GetListenAddress() string {
	return c.Listen
}
func (c *config) GetConfiguratorAddress() string {
	return c.Configurator
}
func (c *config) CreateHandler(ctx context.Context, connectors map[httpservice.ServiceName]*httpservice.InnerService) (httpservice.HttpService, error) {
	return NewRegisterWithForm(c.TrntlAddr, c.TrntlTable, c.TrntlConn, c.MgoAddr, c.MgoColl, c.MgoConn)
}

func (c *config) Close() error {
	c.MgoConn.Close()
	return c.TrntlConn.Close()
}

func main() {
	httpservice.InitNewService(lib.ServiceNameCodesGenerator, false, 5, &config{})
}

// func main() {

// 	servAddr := flag.String("listen", "127.0.0.1:8083", "Service listen address (unix/udp/tcp)")
// 	flgs := &flags{}
// 	flag.StringVar(&flgs.trntlAddr, "trntl-address", "127.0.0.1:3301", "Tarantool listener address (unix/tcp)")
// 	flag.StringVar(&flgs.trntlTable, "trntl-table", "regcodes", "Tarantool table name (string)")
// 	flag.StringVar(&flgs.mgoAddr, "mgo-address", "127.0.0.1", "Mongo listener address (url)")
// 	flag.StringVar(&flgs.mgoCollName, "mgo-coll", "users", "Mongo table name (string)")
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

// 	conf, err := httpservice.NewConfigurator(ctx, lib.ServiceNameRegisterWithForm, *servAddr, httpservice.ServiceName(lib.ServiceNameRegisterWithForm))
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
