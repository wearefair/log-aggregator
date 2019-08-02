package destinations

import "github.com/wearefair/log-aggregator/pkg/types"

// Destination is something we can write logs to
type Destination interface {
	Start(<-chan *types.Record, chan<- types.Cursor)
}
