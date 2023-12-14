package sliceutils

type MapCallback[Input, Output any] func(current Input, index int, all []Input) (Output, error)

func Map[Input, Output any](src []Input, cb MapCallback[Input, Output]) ([]Output, error) {
	out := make([]Output, len(src))
	for i := range src {
		o, err := cb(src[i], i, src)
		if err != nil {
			return nil, err
		}
		out[i] = o
	}

	return out, nil
}
