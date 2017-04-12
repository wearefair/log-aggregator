package json

import (
	"testing"

	"github.com/wearefair/log-aggregator/types"
)

func TestTransform(t *testing.T) {
	// Happy path
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

	// Parse a log field that doesn't have a ts field.
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
	if !transformed.Time.IsZero() {
		t.Errorf("Expected time to be zero")
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != "my wrapped log" {
		t.Errorf("Expected field 'log' to be 'my wrapped log', but got '%s'", val)
	}

	// Parse a log field whose ts field is not a float64
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
	if !transformed.Time.IsZero() {
		t.Errorf("Expected time to be zero")
	}
	if val, ok := transformed.Fields["other"]; !ok || val != "foobar" {
		t.Errorf("Expected field 'other' to be 'foobar', but got '%s'", val)
	}
	if val, ok := transformed.Fields["log"]; !ok || val != "this totally isn't json" {
		t.Errorf("Expected field 'log' to be `this totally ins't json`, but got '%s'", val)
	}

	// Parse a log field that isn't even a string
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
