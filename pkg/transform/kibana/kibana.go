package kibana

import "github.com/wearefair/log-aggregator/pkg/types"

// Transform Adds the @timestamp field that kibana expects to find
func Transform(rec *types.Record) (*types.Record, error) {
	// Time should be something like 	"2017-04-06T20:34:57.961"
	formattedTime := rec.Time.Format("2006-01-02T15:04:05.000")
	rec.Fields["@timestamp"] = formattedTime
	return rec, nil
}
