package utils

import "golang.org/x/exp/constraints"

// Min returns the smaller value between two numbers.
func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return x
}

// Max returns the bigger value between two numbers.
func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

// Abs returns the absolut value of x.
func Abs[T constraints.Signed | constraints.Float](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
