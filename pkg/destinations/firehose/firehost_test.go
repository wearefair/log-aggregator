package firehose

import (
	"testing"

	"github.com/wearefair/log-aggregator/pkg/types"
)

func TestRecordsToBatches(t *testing.T) {
	maxRecords := 2
	maxRecordSize := 50
	maxBatchSize := 80

	records := []*types.Record{
		// These two records should be in one batch
		{
			Cursor: types.Cursor("1"),
			Fields: map[string]interface{}{
				"1234567890": "12345678901234567890",
			},
		},
		{
			Cursor: types.Cursor("2"),
			Fields: map[string]interface{}{
				"1234567890": "09876543210987654321",
			},
		},
		// This record should get truncated
		{
			Cursor: types.Cursor("3"),
			Fields: map[string]interface{}{
				"12345678901234567890": "1234567890abcdefghij1234567890",
			},
		},
		// This record is too large to fit with the truncated record
		{
			Cursor: types.Cursor("4"),
			Fields: map[string]interface{}{
				"12345678901234567890": "12345678901234567890",
			},
		},
	}

	// Expected serialized records
	record1 := "{\"1234567890\":\"12345678901234567890\"}\n"
	record2 := "{\"1234567890\":\"09876543210987654321\"}\n"
	record3 := "{\"12345678901234567890\":\"1234567890abcdefghij1234\n"
	record4 := "{\"12345678901234567890\":\"12345678901234567890\"}\n"

	batches := recordsToBatches(records, maxRecords, maxRecordSize, maxBatchSize)

	if len(batches) != 3 {
		t.Fatalf("Expected 3 batches, but got %d", len(batches))
	}

	// Check the first batch
	if length := len(batches[0].records); length != 2 {
		t.Fatalf("Expected first batch to contain 2 records, but got %d", length)
	}
	if batches[0].cursor != types.Cursor("2") {
		t.Errorf("Expected first batch cursor to be 2, but got %s", batches[0].cursor)
	}
	if val := string(batches[0].records[0].Data); val != record1 {
		t.Errorf("Expected the first record to be serialized to '%s', but got '%s'", record1, val)
	}
	if val := string(batches[0].records[1].Data); val != record2 {
		t.Errorf("Expected the second record to be serialized to '%s', but got '%s'", record2, val)
	}

	// Check the second batch
	if length := len(batches[1].records); length != 1 {
		t.Fatalf("Expected second batch to contain 1 record, but got %d", length)
	}
	if batches[1].cursor != types.Cursor("3") {
		t.Errorf("Expected second batch cursor to be 3, but got %s", batches[1].cursor)
	}
	if val := string(batches[1].records[0].Data); val != record3 {
		t.Errorf("Expected the third record to be serialized to '%s', but got '%s'", record3, val)
	}

	// Check the third batch
	if length := len(batches[2].records); length != 1 {
		t.Fatalf("Expected third batch to contain 1 record, but got %d", length)
	}
	if batches[2].cursor != types.Cursor("4") {
		t.Errorf("Expected third batch cursor to be 4, but got %s", batches[2].cursor)
	}
	if val := string(batches[2].records[0].Data); val != record4 {
		t.Errorf("Expected the fourth record to be serialized to '%s', but got '%s'", record4, val)
	}
}
