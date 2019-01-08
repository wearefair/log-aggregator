package journal

import "github.com/wearefair/log-aggregator/pkg/types"

func Transform(rec *types.Record) (*types.Record, error) {
	// Re-assign MESSAGE field to log
	if val, ok := rec.Fields["MESSAGE"]; ok {
		rec.Fields["log"] = val
		delete(rec.Fields, "MESSAGE")
	}

	// Prefix any fields that start with an underscore with JD
	for k, _ := range rec.Fields {
		if k[0] == '_' {
			rec.Fields["JD"+k] = rec.Fields[k]
			delete(rec.Fields, k)
		}
	}
	return rec, nil
}
