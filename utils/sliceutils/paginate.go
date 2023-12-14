package sliceutils

func Paginate[T any](in []T, offset int, limit int) []T {
	if offset > len(in) {
		offset = len(in)
	}

	end := offset + limit
	if end > len(in) {
		end = len(in)
	}

	return in[offset:end]
}
