package journal

import (
	"testing"

	"github.com/wearefair/log-aggregator/types"
)

func TestTransform(t *testing.T) {
	record := &types.Record{
		Fields: map[string]interface{}{
			"MESSAGE":       "foo",
			"_SYSTEMD_UNIT": "whatever",
		},
	}

	transformed, _ := Transform(record)
	if _, ok := transformed.Fields["MESSAGE"]; ok {
		t.Errorf("Did not expected MESSAGE field to exist")
	}

	if val, ok := transformed.Fields["log"]; !ok || val != "foo" {
		t.Errorf("Expected field log to be 'foo', but got '%s'", val)
	}

	if _, ok := transformed.Fields["_SYSTEMD_UNIT"]; ok {
		t.Errorf("Unexpected field _SYSTEMD_UNIT")
	}

	if val, ok := transformed.Fields["JD_SYSTEMD_UNIT"]; !ok || val != "whatever" {
		t.Errorf("Expected field JD_SYSTEMD_UNIT to be 'whatever', but got '%s'", val)
	}
}
