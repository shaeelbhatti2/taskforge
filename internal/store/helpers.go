package store

import (
	"encoding/json"
)

func encodeJSON(v any) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func decodeJSON(raw string, dest any) {
	if raw == "" {
		return
	}
	_ = json.Unmarshal([]byte(raw), dest)
}

func encodeStringSlice(v []string) string {
	return encodeJSON(v)
}

func decodeStringSlice(raw string) []string {
	var out []string
	decodeJSON(raw, &out)
	return out
}
