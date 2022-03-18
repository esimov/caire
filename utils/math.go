package utils

import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Abs[T constraints.Signed | constraints.Float](i T) T {
	if i < 0 {
		return -i
	}
	return i
}
