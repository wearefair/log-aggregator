package logging

import (
	"os"

	"go.uber.org/zap"
)

// Logger is an instance of *zap.Logger
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

// Error writes an error to the logger
func Error(err error) {
	Logger.Error(err.Error(), zap.Error(err))
}
