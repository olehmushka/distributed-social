package sliceutils

import (
	"fmt"
	"math"

	"github.com/olehmushka/distributed-social/utils/random"
)

func RandomValueOfSlice[T any](randSrc func(float64) (float64, error), in []T) (T, error) {
	var zero T
	if len(in) == 0 {
		return zero, nil
	}
	if len(in) == 1 {
		return in[0], nil
	}

	r, err := randSrc(1)
	if err != nil {
		return zero, err
	}

	return in[int(math.Floor(r*float64(len(in))))], nil
}

func RandomValuesOfSlice[T any](randSrc func(float64) (float64, error), in []T, amount int) ([]T, error) {
	if amount == 0 {
		return []T{}, nil
	}
	if len(in) <= amount {
		return in, nil
	}

	preOut := make(map[int]T)
	for {
		r, err := randSrc(1)
		if err != nil {
			return nil, err
		}

		index := int(math.Floor(r * float64(len(in))))
		preOut[index] = in[index]
		if len(preOut) == amount {
			break
		}
	}

	out := make([]T, 0, amount)
	for _, v := range preOut {
		out = append(out, v)
	}

	return out, nil
}

func RandomValueOfSliceNorm[T any](meanIndex float64, in []T) (T, error) {
	var zero T

	if meanIndex >= float64(len(in)) {
		return zero, fmt.Errorf("impossible pick random value from slice of values by normal distribution (mean index=%v, slice size=%v)", meanIndex, len(in))
	}
	indexF, err := random.GenFloat64NormInRange(0, float64(len(in)-1), 1, float64(meanIndex))
	if err != nil {
		return zero, err
	}

	return in[int(indexF)], nil
}
