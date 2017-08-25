package monitoring

import (
	"github.com/wearefair/service-kit-go/logging"
	"go.uber.org/zap"
)

func Error(err error) {
	logging.Logger().Error(err.Error(), zap.Error(err))
}
