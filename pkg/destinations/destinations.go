package destinations

import "github.com/wearefair/log-aggregator/pkg/types"

type Destination interface {
	Start(<-chan *types.Record, chan<- types.Cursor)
}
