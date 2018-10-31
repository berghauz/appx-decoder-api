package main

import (
	"path/filepath"
	"plugin"
	"strings"

	"github.com/sirupsen/logrus"
)

// Decoder desribe every loaded decoder
type Decoder struct {
	Type        string
	Version     string
	Description string
	Decode      func(string) (interface{}, error) `json:"-"`
}

func (ctx *Context) loadDecoders() {
	// init decoders map
	ctx.DecodingPlugins = make(map[string]Decoder)
	allDecoders, err := filepath.Glob(ctx.DecodersPath + "/*.so")

	if err != nil {
		logger.WithFields(logrus.Fields{"AD_DECODERS_PATH": ctx.DecodersPath}).Fatalf("Can't list decoders dir: %v", err)
	}

	for _, decoder := range allDecoders {
		// try to load decoder
		p, err := plugin.Open(decoder)
		if err != nil {
			logger.WithFields(logrus.Fields{"decoder": decoder}).Fatalf("Can't load decoder: %v", err)
		}
		// import type/name of decoder
		decoderType, err := p.Lookup("Decoder")
		if err != nil {
			logger.WithFields(logrus.Fields{"decoder": decoder}).Fatalf("Can't import type of decoder: %v", err)
		}
		// import decoder version
		decoderVersion, err := p.Lookup("Version")
		if err != nil {
			logger.WithFields(logrus.Fields{"decoder": decoder}).Fatalf("Can't import version of decoder: %v", err)
		}
		// import decoder description
		decoderDescription, err := p.Lookup("Description")
		if err != nil {
			logger.WithFields(logrus.Fields{"decoder": decoder}).Fatalf("Can't import description of decoder: %v", err)
		}
		// import decoder method
		decodeMethod, err := p.Lookup("Decode")
		if err != nil {
			logger.WithFields(logrus.Fields{"decoder": decoder}).Fatalf("Can't import decoder method: %v", err)
		}
		ctx.DecodingPlugins[strings.ToLower(*decoderType.(*string))] = Decoder{Type: strings.ToLower(*decoderType.(*string)), Version: *decoderVersion.(*string), Description: *decoderDescription.(*string), Decode: decodeMethod.(func(string) (interface{}, error))} //decodeMethod.(func(string) (interface{}, error))
	}

	if len(ctx.DecodingPlugins) == 0 {
		logger.Fatalln("No decoders loaded, nothing to do. Bye bye.")
	}

	for decoderType := range ctx.DecodingPlugins {
		logger.Infof("Decoder loaded: %s", decoderType)
	}
}
