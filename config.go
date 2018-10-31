package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
	figure "github.com/common-nighthawk/go-figure"

	"github.com/sirupsen/logrus"
)

// Context contain application runtime variables
type Context struct {
	InventoryURL    string   `env:"AD_INVENTORY_HOST"`
	ListenHost      string   `env:"AD_LISTEN_HOST" envDefault:":3000"`
	PromHost        string   `env:"AD_PROM_HOST" envDefault:":9801"`
	EtcdHosts       []string `env:"AD_ETCD_HOSTS" envSeparator:","`
	DecodersPath    string   `env:"AD_DECODERS_PATH"`
	LoggerLevel     string   `env:"AD_LOG_LEVEL" envDefault:"info"`
	LogPath         string   `env:"AD_LOG_PATH" envDefault:"stdout"`
	DecodingPlugins map[string]Decoder
}

var (
	help       *bool
	logger     *logrus.Logger
	version    = "UNDEFINED"
	buildstamp = "UNDEFINED"
	githash    = "UNDEFINED"
)

func init() {
	// funny shit
	myFigure := figure.NewFigure("appx decoder", "ogre", true)
	myFigure.Print()
	fmt.Printf("\n%s\nGIT Commit Hash: %s\nBuild Time: %s\n\n", version, githash, buildstamp)

	help = flag.Bool("h", false, "print defaults")
	flag.Parse()

	if *help {
		fmt.Printf("Available env variables:\nAD_INVENTORY_HOST: inventory host:port\nAD_LISTEN_HOST: application listen on host:port\n")
		os.Exit(0)
	}

	logger = logrus.New()

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05-07:00",
		DisableColors:   false,
	}
	logger.Formatter = formatter

	logger.SetLevel(logrus.InfoLevel)
	logger.Out = os.Stdout

}

// createContext load configuration from env and perform application initialize
func createContext() *Context {

	ctx := Context{}

	err := env.Parse(&ctx)
	if err != nil {
		logger.Fatalf("%+v", err)
	}

	// reconfigure logger if needed
	if ctx.LoggerLevel != "info" {
		ll, err := logrus.ParseLevel(ctx.LoggerLevel)
		if err != nil {
			ll = logrus.InfoLevel
		}
		logger.SetLevel(ll)
	}

	logger.Infof("Log level is %s", logger.Level)

	if ctx.LogPath != "stdout" {
		// prepare log file
		handle, err := os.OpenFile(ctx.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.WithFields(logrus.Fields{"AD_LOG_PATH": ctx.LogPath}).Fatalf("%+v", err)
		}
		logger.Out = handle
	} else {
		logger.Out = os.Stdout
	}

	logger.Infof("Log output is %s", ctx.LogPath)

	// check if inventory endpoint is configured
	if ctx.InventoryURL == "" {
		logger.Fatalln("AD_INVENTORY_HOST not set, terminating")
	}

	// check drivers path availability
	if info, err := os.Stat(ctx.DecodersPath); os.IsNotExist(err) {
		logger.WithFields(logrus.Fields{"AD_DRIVERS_PATH": ctx.DecodersPath}).Fatalln("Not exists")
	} else {
		mode := info.Mode()
		//if !mode.IsDir() && mode.Perm()
		logger.Infof("AD_DRIVERS_PATH permission is %s", mode&os.ModePerm)
	}

	ctx.loadDecoders()

	return &ctx
}
