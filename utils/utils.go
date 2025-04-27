package utils

import (
	"slices"

	"golang.org/x/exp/constraints"
)

// Min returns the slowest value of the provided parameters.
func Min[T constraints.Ordered](values ...T) T {
	var acc T = values[0]

	for _, v := range values {
		if v < acc {
			acc = v
		}
	}
	return acc
}

// Max returns the biggest value of the provided parameters.
func Max[T constraints.Ordered](values ...T) T {
	var acc T = values[0]

	for _, v := range values {
		if v > acc {
			acc = v
		}
	}
	return acc
}

// Abs returns the absolut value of x.
func Abs[T constraints.Signed | constraints.Float](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// Contains returns true if a value is available in the collection.
func Contains[T comparable](slice []T, value T) bool {
	return slices.Contains(slice, value)
}
