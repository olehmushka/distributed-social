package sliceutils

import (
	"github.com/google/uuid"
)

// Unique primitive values slice.
func Unique[T uuid.UUID | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](sl []T) []T {
	if len(sl) <= 1 {
		return sl
	}

	var (
		keys = make(map[T]bool)
		list = make([]T, 0)
	)
	for _, entry := range sl {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// UniqueNestedSlice primitive values slice.
func UniqueNestedSlice[T uuid.UUID | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](sl [][]T, index int) [][]T {
	if len(sl) <= 1 {
		return sl
	}

	var (
		keys = make(map[T]bool)
		list = make([][]T, 0)
	)
	for _, entry := range sl {
		if _, value := keys[entry[index]]; !value {
			keys[entry[index]] = true
			list = append(list, entry)
		}
	}
	return list
}
