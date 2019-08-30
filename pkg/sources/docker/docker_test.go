package docker

import (
	"github.com/wearefair/log-aggregator/pkg/sources"
)

var _ sources.Source = &Docker{}
