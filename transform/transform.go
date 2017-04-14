package transform

import "github.com/wearefair/log-aggregator/types"

type Transformer func(rec *types.Record) (*types.Record, error)
