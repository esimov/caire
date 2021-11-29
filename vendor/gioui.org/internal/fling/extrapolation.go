// SPDX-License-Identifier: Unlicense OR MIT

package fling

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// Extrapolation computes a 1-dimensional velocity estimate
// for a set of timestamped points using the least squares
// fit of a 2nd order polynomial. The same method is used
// by Android.
type Extrapolation struct {
	// Index into points.
	idx int
	// Circular buffer of samples.
	samples   []sample
	lastValue float32
	// Pre-allocated cache for samples.
	cache [historySize]sample

	// Filtered values and times
	values [historySize]float32
	times  [historySize]float32
}

type sample struct {
	t time.Duration
	v float32
}

type matrix struct {
	rows, cols int
	data       []float32
}

type Estimate struct {
	Velocity float32
	Distance float32
}

type coefficients [degree + 1]float32

const (
	degree       = 2
	historySize  = 20
	maxAge       = 100 * time.Millisecond
	maxSampleGap = 40 * time.Millisecond
)

// SampleDelta adds a relative sample to the estimation.
func (e *Extrapolation) SampleDelta(t time.Duration, delta float32) {
	val := delta + e.lastValue
	e.Sample(t, val)
}

// Sample adds an absolute sample to the estimation.
func (e *Extrapolation) Sample(t time.Duration, val float32) {
	e.lastValue = val
	if e.samples == nil {
		e.samples = e.cache[:0]
	}
	s := sample{
		t: t,
		v: val,
	}
	if e.idx == len(e.samples) && e.idx < cap(e.samples) {
		e.samples = append(e.samples, s)
	} else {
		e.samples[e.idx] = s
	}
	e.idx++
	if e.idx == cap(e.samples) {
		e.idx = 0
	}
}

// Velocity returns an estimate of the implied velocity and
// distance for the points sampled, or zero if the estimation method
// failed.
func (e *Extrapolation) Estimate() Estimate {
	if len(e.samples) == 0 {
		return Estimate{}
	}
	values := e.values[:0]
	times := e.times[:0]
	first := e.get(0)
	t := first.t
	// Walk backwards collecting samples.
	for i := 0; i < len(e.samples); i++ {
		p := e.get(-i)
		age := first.t - p.t
		if age >= maxAge || t-p.t >= maxSampleGap {
			// If the samples are too old or
			// too much time passed between samples
			// assume they're not part of the fling.
			break
		}
		t = p.t
		values = append(values, first.v-p.v)
		times = append(times, float32((-age).Seconds()))
	}
	coef, ok := polyFit(times, values)
	if !ok {
		return Estimate{}
	}
	dist := values[len(values)-1] - values[0]
	return Estimate{
		Velocity: coef[1],
		Distance: dist,
	}
}

func (e *Extrapolation) get(i int) sample {
	idx := (e.idx + i - 1 + len(e.samples)) % len(e.samples)
	return e.samples[idx]
}

// fit computes the least squares polynomial fit for
// the set of points in X, Y. If the fitting fails
// because of contradicting or insufficient data,
// fit returns false.
func polyFit(X, Y []float32) (coefficients, bool) {
	if len(X) != len(Y) {
		panic("X and Y lengths differ")
	}
	if len(X) <= degree {
		// Not enough points to fit a curve.
		return coefficients{}, false
	}

	// Use a method similar to Android's VelocityTracker.cpp:
	// https://android.googlesource.com/platform/frameworks/base/+/56a2301/libs/androidfw/VelocityTracker.cpp
	// where all weights are 1.

	// First, expand the X vector to the matrix A in column-major order.
	A := newMatrix(degree+1, len(X))
	for i, x := range X {
		A.set(0, i, 1)
		for j := 1; j < A.rows; j++ {
			A.set(j, i, A.get(j-1, i)*x)
		}
	}

	Q, Rt, ok := decomposeQR(A)
	if !ok {
		return coefficients{}, false
	}
	// Solve R*B = Qt*Y for B, which is then the polynomial coefficients.
	// Since R is upper triangular, we can proceed from bottom right to
	// upper left.
	// https://en.wikipedia.org/wiki/Non-linear_least_squares
	var B coefficients
	for i := Q.rows - 1; i >= 0; i-- {
		B[i] = dot(Q.col(i), Y)
		for j := Q.rows - 1; j > i; j-- {
			B[i] -= Rt.get(i, j) * B[j]
		}
		B[i] /= Rt.get(i, i)
	}
	return B, true
}

// decomposeQR computes and returns Q, Rt where Q*transpose(Rt) = A, if
// possible. R is guaranteed to be upper triangular and only the square
// part of Rt is returned.
func decomposeQR(A *matrix) (*matrix, *matrix, bool) {
	// Gram-Schmidt QR decompose A where Q*R = A.
	// https://en.wikipedia.org/wiki/Gram%E2%80%93Schmidt_process
	Q := newMatrix(A.rows, A.cols)  // Column-major.
	Rt := newMatrix(A.rows, A.rows) // R transposed, row-major.
	for i := 0; i < Q.rows; i++ {
		// Copy A column.
		for j := 0; j < Q.cols; j++ {
			Q.set(i, j, A.get(i, j))
		}
		// Subtract projections. Note that int the projection
		//
		// proju a = <u, a>/<u, u> u
		//
		// the normalized column e replaces u, where <e, e> = 1:
		//
		// proje a = <e, a>/<e, e> e = <e, a> e
		for j := 0; j < i; j++ {
			d := dot(Q.col(j), Q.col(i))
			for k := 0; k < Q.cols; k++ {
				Q.set(i, k, Q.get(i, k)-d*Q.get(j, k))
			}
		}
		// Normalize Q columns.
		n := norm(Q.col(i))
		if n < 0.000001 {
			// Degenerate data, no solution.
			return nil, nil, false
		}
		invNorm := 1 / n
		for j := 0; j < Q.cols; j++ {
			Q.set(i, j, Q.get(i, j)*invNorm)
		}
		// Update Rt.
		for j := i; j < Rt.cols; j++ {
			Rt.set(i, j, dot(Q.col(i), A.col(j)))
		}
	}
	return Q, Rt, true
}

func norm(V []float32) float32 {
	var n float32
	for _, v := range V {
		n += v * v
	}
	return float32(math.Sqrt(float64(n)))
}

func dot(V1, V2 []float32) float32 {
	var d float32
	for i, v1 := range V1 {
		d += v1 * V2[i]
	}
	return d
}

func newMatrix(rows, cols int) *matrix {
	return &matrix{
		rows: rows,
		cols: cols,
		data: make([]float32, rows*cols),
	}
}

func (m *matrix) set(row, col int, v float32) {
	if row < 0 || row >= m.rows {
		panic("row out of range")
	}
	if col < 0 || col >= m.cols {
		panic("col out of range")
	}
	m.data[row*m.cols+col] = v
}

func (m *matrix) get(row, col int) float32 {
	if row < 0 || row >= m.rows {
		panic("row out of range")
	}
	if col < 0 || col >= m.cols {
		panic("col out of range")
	}
	return m.data[row*m.cols+col]
}

func (m *matrix) col(c int) []float32 {
	return m.data[c*m.cols : (c+1)*m.cols]
}

func (m *matrix) approxEqual(m2 *matrix) bool {
	if m.rows != m2.rows || m.cols != m2.cols {
		return false
	}
	const epsilon = 0.00001
	for row := 0; row < m.rows; row++ {
		for col := 0; col < m.cols; col++ {
			d := m2.get(row, col) - m.get(row, col)
			if d < -epsilon || d > epsilon {
				return false
			}
		}
	}
	return true
}

func (m *matrix) transpose() *matrix {
	t := &matrix{
		rows: m.cols,
		cols: m.rows,
		data: make([]float32, len(m.data)),
	}
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			t.set(j, i, m.get(i, j))
		}
	}
	return t
}

func (m *matrix) mul(m2 *matrix) *matrix {
	if m.rows != m2.cols {
		panic("mismatched matrices")
	}
	mm := &matrix{
		rows: m.rows,
		cols: m2.cols,
		data: make([]float32, m.rows*m2.cols),
	}
	for i := 0; i < mm.rows; i++ {
		for j := 0; j < mm.cols; j++ {
			var v float32
			for k := 0; k < m.rows; k++ {
				v += m.get(k, j) * m2.get(i, k)
			}
			mm.set(i, j, v)
		}
	}
	return mm
}

func (m *matrix) String() string {
	var b strings.Builder
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			v := m.get(i, j)
			b.WriteString(strconv.FormatFloat(float64(v), 'g', -1, 32))
			b.WriteString(", ")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (c coefficients) approxEqual(c2 coefficients) bool {
	const epsilon = 0.00001
	for i, v := range c {
		d := v - c2[i]
		if d < -epsilon || d > epsilon {
			return false
		}
	}
	return true
}
