package docker

import (
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/wearefair/log-aggregator/pkg/types"
)

type container struct {
	id      string
	lastLog time.Time
	rdr     io.ReadCloser
	out     chan<- *types.Record
}

const (
	bufSizeInBytes = 1024 * 10 // 10K

	eventsToSubscribe = events.ContainerEventType
	eventsCreation    = "container create"
)

var eventsDestruction = []string{"container die", "container kill", "container stop"}

func (container *container) read() {
	buf := make([]byte, bufSizeInBytes)
	for {
		n, err := container.rdr.Read(buf)
		if n == 0 {
			break
		}
		if err == io.EOF {
			break
		}
		container.lastLog = time.Now()
		container.out <- container.transform(buf)
	}
}

func (container *container) stop() error {
	return container.rdr.Close()
}

func (container *container) transform(data []byte) *types.Record {
	fields := make(map[string]interface{})
	return &types.Record{
		Time:   time.Now(),
		Cursor: types.Cursor(string(container.lastLog.Unix())),
		Fields: fields,
	}
}

func isContainerCreation(msg events.Message) (string, bool) {
	if msg.Type != eventsToSubscribe {
		return "", false
	}
	if strings.Contains(msg.Action, eventsCreation) {
		return msg.Actor.ID, true
	}
	return "", false
}

func isContainerDestruction(msg events.Message) (string, bool) {
	if msg.Type != eventsToSubscribe {
		return "", false
	}
	for _, destructionEvent := range eventsDestruction {
		if strings.Contains(msg.Action, destructionEvent) {
			return msg.Actor.ID, true
		}
	}
	return "", false
}
