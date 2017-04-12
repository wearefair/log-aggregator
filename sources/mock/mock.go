package mock

import (
	"time"

	"github.com/wearefair/log-aggregator/types"
)

type Client struct {
	ticker   *time.Ticker
	interval time.Duration
}

func New(interval time.Duration) *Client {
	return &Client{
		interval: interval,
	}
}

func (c *Client) Start(out chan<- *types.Record) {
	c.ticker = time.NewTicker(c.interval)
	go func() {
		for {
			t, ok := <-c.ticker.C
			if !ok {
				return
			}
			out <- &types.Record{
				Time:   t,
				Cursor: types.Cursor(t.String()),
				Fields: map[string]interface{}{
					"MESSAGE":           `{"log":"my fake log","ts":1492015752.123456789,"hello":"field"}`,
					"CONTAINER_NAME":    "k8s_containername.containerhash_contract-service-2957857213-vztuq_default_poduuid_abcd1234",
					"CONTAINER_ID_FULL": "mycontainerid",
				},
			}
		}
	}()
}

func (c *Client) Stop() {
	c.ticker.Stop()
}
