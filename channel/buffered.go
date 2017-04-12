package channel

import (
	"time"

	"github.com/wearefair/log-aggregator/types"
)

// Returns a channel that will flush a list of Record objects either every "interval", or when the buffer
// has reached the specified size, whichever comes first.
//
// This is used to buffer logs in destinations (ex: firehose) where we want to send
// batches of logs at a time, but we also want the logs to be delivered in a timely manner.
func NewBufferedChannel(size int, interval time.Duration, in <-chan *types.Record) <-chan []*types.Record {
	index := 0
	buffer := make([]*types.Record, size)
	ticker := time.NewTicker(interval)
	out := make(chan []*types.Record)
	go func() {
		for {
			select {
			case record, ok := <-in:
				if !ok {
					close(out)
					ticker.Stop()
					return
				}
				buffer[index] = record
				index++
				if index == size {
					out <- buffer
					buffer = make([]*types.Record, size)
					index = 0
				}

			case <-ticker.C:
				if index != 0 {
					out <- buffer[0:index]
					buffer = make([]*types.Record, size)
					index = 0
				}
			}
		}
	}()
	return out
}
