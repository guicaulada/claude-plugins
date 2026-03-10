package events

import (
	"encoding/json"
	"time"
)

// marshalAttrs safely serializes a map to JSON for span event storage.
func marshalAttrs(attrs map[string]string) string {
	data, err := json.Marshal(attrs)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// currentTimestamp returns the current time as UnixNano.
func currentTimestamp() int64 {
	return time.Now().UnixNano()
}
