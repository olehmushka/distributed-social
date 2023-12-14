package sliceutils

func Pop[T any](in []T, index int) []T {
	return append(in[:index], in[index+1:]...)
}
