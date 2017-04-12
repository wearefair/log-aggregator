package types

import "time"

type Cursor string

type Record struct {
	Time   time.Time
	Cursor Cursor
	Fields map[string]interface{}
}
