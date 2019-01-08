// Package stdout provides a simple destination that prints all records to STDOUT.
// It is primarily used for debugging and development.
package stdout

import (
	"encoding/json"
	"fmt"

	"github.com/wearefair/log-aggregator/logging"
	"github.com/wearefair/log-aggregator/types"
)

type Client struct{}

func New() *Client {
	return &Client{}
}

func (c *Client) Start(records <-chan *types.Record, progress chan<- types.Cursor) {
	go func() {
		for {
			record, open := <-records
			if !open {
				logging.Logger.Warn("record channel was unexpectedly closed")
				return
			}
			jsonBytes, err := json.Marshal(record.Fields)
			if err != nil {
				logging.Error(err)
				break
			}
			fmt.Println(string(jsonBytes))
			progress <- record.Cursor
		}
	}()
}
