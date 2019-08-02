package json

import (
	"testing"

	"github.com/wearefair/log-aggregator/pkg/types"
)

func TestTransform(t *testing.T) {
	record := &types.Record{
		Fields: map[string]interface{}{
			"log":   "{\"log\":\"my wrapped log\",\"ts\":1487349663.5884562}",
			"other": "foobar",
		},
	}
	transformed, err := Transform(record)
	if err != nil {
		t.Fatal(err)
	}
	checkHappyPath(t, transformed)

	record = &types.Record{
		Fields: map[string]interface{}{
			"log":   "{\"log\":\"my wrapped log\"}",
			"other": "foobar",
		},
	}
	transformed, err = Transform(record)
	if err != nil {
		t.Fatal(err)
	}
	checkLogFieldThatIsNoTSField(t, transformed)

	record = &types.Record{
		Fields: map[string]interface{}{
			"log":   "{\"log\":\"my wrapped log\",\"ts\":\"not a time\"}",
			"other": "foobar",
		},
	}
	transformed, err = Transform(record)
	if err != nil {
		t.Fatal(err)
	}
	checkLogFieldThatIsNotFloat64(t, transformed)
	// Parse a log field that isn't json
	record = &types.Record{
		Fields: map[string]interface{}{
			"log":   "this totally isn't json",
			"other": "foobar",
		},
	}
	transformed, err = Transform(record)
	if err != nil {
		t.Fatal(err)
	}
	checkLogFieldThatIsNotJSON(t, transformed)

	record = &types.Record{
		Fields: map[string]interface{}{
			"log":   12345,
			"other": "foobar",
		},
	}
	transformed, err = Transform(record)
	if err != nil {
		t.Fatal(err)
	}
	checkLogFieldThatIsNotAString(t, transformed)
}

func checkHappyPath(t *testing.T, transformed *types.Record) {
	if transformed.Time.Unix() != 1487349663 || transformed.Time.Nanosecond() != 588456153 {
		t.Errorf("Expected record time to be 1487349663.588456153, but got %d.%d",
			transformed.Time.Unix(), transformed.Time.Nanosecond())
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != "my wrapped log" {
		t.Errorf("Expected field 'log' to be 'my wrapped log', but got '%s'", val)
	}
	if val, ok := transformed.Fields["ts"]; !ok || val != 1487349663.5884562 {
		t.Errorf("Expected field 'ts' to be %f, but got %f", 1487349663.5884562, val)
	}
}

func checkLogFieldThatIsNoTSField(t *testing.T, transformed *types.Record) {
	if !transformed.Time.IsZero() {
		t.Errorf("Expected time to be zero")
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != "my wrapped log" {
		t.Errorf("Expected field 'log' to be 'my wrapped log', but got '%s'", val)
	}
}

func checkLogFieldThatIsNotFloat64(t *testing.T, transformed *types.Record) {
	if !transformed.Time.IsZero() {
		t.Errorf("Expected time to be zero")
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != "my wrapped log" {
		t.Errorf("Expected field 'log' to be 'my wrapped log', but got '%s'", val)
	}
	if val, ok := transformed.Fields["ts"]; !ok || val != "not a time" {
		t.Errorf("Expected field 'ts' to be 'not a time', but got '%s'", val)
	}
}

func checkLogFieldThatIsNotJSON(t *testing.T, transformed *types.Record) {
	if !transformed.Time.IsZero() {
		t.Errorf("Expected time to be zero")
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != "this totally isn't json" {
		t.Errorf("Expected field 'log' to be `this totally ins't json`, but got '%s'", val)
	}
}

func checkLogFieldThatIsNotAString(t *testing.T, transformed *types.Record) {
	if !transformed.Time.IsZero() {
		t.Errorf("Expected time to be zero")
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != 12345 {
		t.Errorf("Expected field 'log' to be 12345, but got '%s'", val)
	}
}
