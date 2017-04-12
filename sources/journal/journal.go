package journal

import (
	"math/big"
	"strconv"
	"time"

	"github.com/wearefair/log-aggregator/types"
)

const SD_JOURNAL_FIELD_SOURCE_REALTIME_TIMESTAMP = "_SOURCE_REALTIME_TIMESTAMP"

type ClientConfig struct {
	JournalDirectory string
	Cursor           types.Cursor
}

func (c *Client) Stop() {
	c.shutdown = true
}

func entryToRecord(entry *JournalEntry) *types.Record {
	fields := make(map[string]interface{})
	entryTime := entryToTime(entry)

	// Remove fields that we don't want to forward
	for _, field := range omitFields {
		delete(entry.Fields, field)
	}

	// Copy entry fields to record fields
	for k, _ := range entry.Fields {
		fields[k] = entry.Fields[k]
	}

	return &types.Record{
		Time:   entryTime,
		Cursor: types.Cursor(entry.Cursor),
		Fields: fields,
	}
}

func entryToTime(entry *JournalEntry) time.Time {
	if sourceTimestamp, ok := entry.Fields[SD_JOURNAL_FIELD_SOURCE_REALTIME_TIMESTAMP]; ok {
		secondsStr := sourceTimestamp[0 : len(sourceTimestamp)-6]
		msecondsStr := sourceTimestamp[len(sourceTimestamp)-6:]
		seconds, serr := strconv.ParseInt(secondsStr, 10, 64)
		mseconds, mserr := strconv.ParseInt(msecondsStr, 10, 64)

		// If we don't get errors parsing the time, then return a new time instance.
		// Otherwise fall through and use entry.RealtimeTimestamp
		if serr == nil && mserr == nil {
			return time.Unix(seconds, mseconds*int64(time.Microsecond))
		}
	}

	seconds := big.NewInt(0)
	mseconds := big.NewInt(0)
	seconds.SetUint64(entry.RealtimeTimestamp)
	mseconds.SetUint64(entry.RealtimeTimestamp)
	mod := big.NewInt(int64(time.Millisecond))
	seconds.Div(seconds, mod)
	mseconds.Mod(mseconds, mod)
	return time.Unix(seconds.Int64(), mseconds.Int64()*int64(time.Microsecond))
}
