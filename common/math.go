package common

import (
	"slices"

	"golang.org/x/exp/constraints"
)

type Integer interface {
	constraints.Signed | constraints.Unsigned
}

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

func Median[T constraints.Integer | constraints.Float](min, max T) T {
	return min + (max-min)/2
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

// Clamp is a generic clamp function for integers.
func Clamp[T Integer](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
