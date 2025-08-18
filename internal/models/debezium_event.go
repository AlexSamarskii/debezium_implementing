package models

import "encoding/json"

type DebeziumEvent struct {
	Payload struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
		Op     string          `json:"op"` // c=insert, u=update, d=delete, r=snapshot
		TsMs   int64           `json:"ts_ms"`
	} `json:"payload"`
}
