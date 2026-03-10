package events

import "encoding/json"

// marshalAttrs safely serializes a map to JSON for span event storage.
func marshalAttrs(attrs map[string]string) string {
	data, err := json.Marshal(attrs)
	if err != nil {
		return "{}"
	}
	return string(data)
}
