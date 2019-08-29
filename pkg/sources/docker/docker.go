package docker

import (
	"github.com/docker/docker/client"
	"github.com/wearefair/log-aggregator/pkg/types"
)

// NewDocker returns a new Docker log reader
func NewDocker() *Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	return &Docker{
		files: make(map[string]*file),
	}
}

// Docker reads Docker log files in JSON format
type Docker struct {
	client *client.Client
	files  map[string]*file
}

// Start statts the Docker reader
func (file *Docker) Start(out chan<- *types.Record) {

}

// Stop stops the Docker reader
func (file *Docker) Stop() {

}
