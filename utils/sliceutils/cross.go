package sliceutils

func GetCrossOfSlices[T any](s1, s2 []T, compare func(T, T) bool) []T {
	if len(s1) == 0 || len(s2) == 0 {
		return []T{}
	}

	out := make([]T, 0, len(s1))
	for _, el := range s1 {
		for _, innerEl := range s2 {
			if compare(el, innerEl) {
				out = append(out, innerEl)
			}
		}
	}

	return out
}
