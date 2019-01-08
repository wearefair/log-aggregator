package kibana

import (
	"testing"
	"time"

	"github.com/wearefair/log-aggregator/pkg/types"
)

func TestTransform(t *testing.T) {
	record := &types.Record{
		Time:   time.Date(2017, 4, 3, 15, 32, 45, 120456789, time.UTC),
		Fields: map[string]interface{}{},
	}

	transformed, _ := Transform(record)

	expectedTimeString := "2017-04-03T15:32:45.120"
	if val, ok := transformed.Fields["@timestamp"]; !ok || val != expectedTimeString {
		t.Errorf("Expected @timestamp to be '%s', but got '%s'", expectedTimeString, val)
	}
}
