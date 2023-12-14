package sliceutils

import (
	"github.com/google/uuid"
)

func Includes[T uuid.UUID | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](values []T, v T) bool {
	for _, val := range values {
		if val == v {
			return true
		}
	}

	return false
}
