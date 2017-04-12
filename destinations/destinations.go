package destinations

import "github.com/wearefair/log-aggregator/types"

type Destination interface {
	Start(<-chan *types.Record, chan<- types.Cursor)
}
