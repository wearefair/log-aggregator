package journal

import "github.com/wearefair/log-aggregator/pkg/types"

var omitFields = []string{
	SD_JOURNAL_FIELD_SOURCE_REALTIME_TIMESTAMP,
}

type JournalEntry struct {
	Fields             map[string]string
	Cursor             string
	RealtimeTimestamp  uint64
	MonotonicTimestamp uint64
}

type Client struct {
	shutdown bool
	out      chan<- *types.Record
}

func New(conf ClientConfig) (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Start(out chan<- *types.Record) {
	c.out = out
}
