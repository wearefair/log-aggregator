package sources

import "github.com/wearefair/log-aggregator/pkg/types"

// Source is anything that can produce logs, e.g. Journald
type Source interface {
	Start(chan<- *types.Record)
	Stop()
}
