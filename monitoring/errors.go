package monitoring

import (
	"github.com/wearefair/go-http-kit/logging"
	"go.uber.org/zap"
)

func Error(err error) {
	logging.Logger().Error(err.Error(), zap.Error(err))
}
