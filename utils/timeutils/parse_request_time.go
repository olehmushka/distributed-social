package timeutils

import "time"

func ParseRequestTime(v string) time.Time {
	var zero time.Time
	if out, err := time.Parse(time.RFC3339, v); err == nil {
		return out
	}
	if out, err := time.Parse(time.RFC3339Nano, v); err == nil {
		return out
	}
	if out, err := time.Parse(time.DateTime, v); err == nil {
		return out
	}
	if out, err := time.Parse(time.DateOnly, v); err == nil {
		return out
	}

	return zero
}
