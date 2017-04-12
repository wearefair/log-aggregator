package sources

import "github.com/wearefair/log-aggregator/types"

type Source interface {
	Start(chan<- *types.Record)
	Stop()
}
