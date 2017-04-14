package json

import (
	"encoding/json"
	"math"
	"time"

	"github.com/wearefair/log-aggregator/types"
)

func Transform(rec *types.Record) (*types.Record, error) {
	if log, ok := rec.Fields["log"]; ok {
		if logString, ok := log.(string); ok {
			var parsed map[string]interface{}
			err := json.Unmarshal([]byte(logString), &parsed)
			// We ignore parse errors, as a log could be a format other than json
			if err != nil {
				return rec, nil
			}

			// Copy parsed json fields onto root of record fields.
			for k, _ := range parsed {
				rec.Fields[k] = parsed[k]
			}

			// If the "ts" field is present, attempt to renew the record time
			if ts, ok := rec.Fields["ts"]; ok {
				if tsFloat, ok := ts.(float64); ok {
					seconds, subseconds := math.Modf(tsFloat)
					// support up to 9 digits of accuracy, not super concerned with rounding errors
					rec.Time = time.Unix(int64(seconds), int64(subseconds*float64(time.Second)))
				}
			}
		}
	}
	return rec, nil
}
