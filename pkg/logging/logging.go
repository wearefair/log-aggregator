package logging

import (
	"os"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	var err error
	if os.Getenv("ENV") == "production" {
		conf := zap.NewProductionConfig()
		conf.EncoderConfig.MessageKey = "log"
		Logger, err = conf.Build()
	} else {
		Logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
}

func Error(err error) {
	Logger.Error(err.Error(), zap.Error(err))
}
