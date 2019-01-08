package channel

import (
	"testing"
	"time"

	"github.com/wearefair/log-aggregator/pkg/types"
)

func TestBufferedChannel(t *testing.T) {
	in := make(chan *types.Record, 10)
	out := NewBufferedChannel(2, time.Millisecond*200, in)

	// Write three records in, and expect first 2 to be published almost instantly
	in <- &types.Record{Cursor: types.Cursor("1")}
	in <- &types.Record{Cursor: types.Cursor("2")}
	in <- &types.Record{Cursor: types.Cursor("3")}

	// Expect the first two records to be published.
	select {
	case <-time.After(time.Millisecond * 100):
		t.Fatalf("Expected channel to publish slice of 2 records, but timed out after 100 milliseconds")

	case records, ok := <-out:
		if !ok {
			t.Fatalf("Expected list of 2 records, but channel was closed")
		}
		if length := len(records); length != 2 {
			t.Fatalf("Expected list of 2 records but got %d", length)
		}
		if records[0].Cursor != types.Cursor("1") {
			t.Errorf("Expected first record cursor to be 1, but got %s", records[0].Cursor)
		}
		if records[1].Cursor != types.Cursor("2") {
			t.Errorf("Expected first record cursor to be 2, but got %s", records[0].Cursor)
		}
	}

	// Expect the third record to be published after a total of 5 seconds
	select {
	case <-time.After(time.Millisecond * 300):
		t.Fatalf("Expected channel to publish slice of 1 record, but timed out after 300 milliseconds")

	case records, ok := <-out:
		if !ok {
			t.Fatalf("Expected list of 1 record, but channel was closed")
		}
		if length := len(records); length != 1 {
			t.Fatalf("Expected list of 1 record but got %d", length)
		}
		if records[0].Cursor != types.Cursor("3") {
			t.Errorf("Expected first record cursor to be 3, but got %s", records[0].Cursor)
		}
	}

	// Close the in channel, and expect that the out channel gets closed as well
	close(in)
	select {
	case <-time.After(time.Millisecond * 200):
		t.Fatalf("Expected channel to be closed, but timed out after 200ms")

	case _, ok := <-out:
		if ok {
			t.Errorf("Expected channel to be closed, but it wasn't")
		}
	}
}
