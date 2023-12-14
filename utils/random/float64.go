package random

import (
	"fmt"

	expRand "golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

func GenFloat64InRange(min, max float64) (float64, error) {
	if min >= max {
		return 0, fmt.Errorf("impossible to generate random float64 for max (%v) <= min (%v)", max, min)
	}

	rInt, err := GenIntInRange(0, 100)
	if err != nil {
		return 0, fmt.Errorf("failed to generate int for float64 generation (err=%w)", err)
	}
	s := expRand.NewSource(uint64(rInt))

	return min + expRand.New(s).Float64()*(max-min), nil
}

func GenFloat64(max float64) (float64, error) {
	return GenFloat64InRange(0, max)
}

// GenFloat64Norm generate random float64 with norm
//
// stdDev -standart deviation - σ^2; default = 1
//
// mean - μ (In probability theory, the expected value is a generalization of the weighted average.); default = 0
func GenFloat64Norm(stdDev, mean float64) (float64, error) {
	rInt, err := GenIntInRange(0, 100)
	if err != nil {
		return 0, err
	}
	s := expRand.NewSource(uint64(rInt))

	dist := distuv.Normal{
		Mu:    mean,   // Mean of the normal distribution
		Sigma: stdDev, // Standard deviation of the normal distribution
		Src:   s,
	}

	return dist.Rand(), nil
}

func GenFloat64NormInRange(min, max, stdDev, mean float64) (float64, error) {
	if min >= max {
		return 0, fmt.Errorf("impossible to generate random float64 for max (%v) <= min (%v)", max, min)
	}

	out, err := genFloat64NormInRange(min, max, stdDev, mean, 10)
	if err != nil {
		return 0, err
	}

	return out, nil
}

func genFloat64NormInRange(min, max, stdDev, mean float64, count int) (float64, error) {
	var multiplier float64 = 10

	min *= multiplier
	max *= multiplier
	mean *= multiplier

	r, err := GenFloat64Norm(stdDev, mean)
	if err != nil {
		return 0, err
	}

	if r < min {
		if count == 0 {
			return min / multiplier, nil
		}
		return genFloat64NormInRange(min/multiplier, max/multiplier, stdDev, mean/multiplier, count-1)
	}
	if r > max {
		if count == 0 {
			return max / multiplier, nil
		}
		return genFloat64NormInRange(min/multiplier, max/multiplier, stdDev, mean/multiplier, count-1)
	}

	return r / multiplier, nil
}
