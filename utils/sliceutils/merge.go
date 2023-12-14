package sliceutils

func Merge[T any](in ...[]T) []T {
	out := make([]T, 0)
	for _, els := range in {
		out = append(out, els...)
	}

	return out
}
