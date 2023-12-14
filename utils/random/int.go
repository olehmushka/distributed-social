package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenIntInRange(min int, max int) (int, error) {
	if min >= max {
		return 0, fmt.Errorf("impossible to generate random int for max (%v) <= min (%v)", max, min)
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		return 0, err
	}

	return int(int64(min) + n.Int64()), nil
}

func GenInt(max int) (int, error) {
	out, err := GenIntInRange(0, max)
	if err != nil {
		return 0, err
	}

	return out, nil
}
