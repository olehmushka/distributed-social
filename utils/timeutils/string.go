package timeutils

import (
	"time"
)

func TimeToString(in time.Time) string {
	return in.Format(DefaultTimeFormat)
}

func StringToTime(in string) (time.Time, error) {
	out, err := time.Parse(DefaultTimeFormat, in)
	if err != nil {
		return time.Time{}, err
	}

	return out, nil
}
