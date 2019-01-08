package sources

import "github.com/wearefair/log-aggregator/pkg/types"

type Source interface {
	Start(chan<- *types.Record)
	Stop()
}
