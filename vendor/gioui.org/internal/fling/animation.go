// SPDX-License-Identifier: Unlicense OR MIT

package fling

import (
	"math"
	"runtime"
	"time"

	"gioui.org/unit"
)

type Animation struct {
	// Current offset in pixels.
	x float32
	// Initial time.
	t0 time.Time
	// Initial velocity in pixels pr second.
	v0 float32
}

var (
	// Pixels/second.
	minFlingVelocity = unit.Dp(50)
	maxFlingVelocity = unit.Dp(8000)
)

const (
	thresholdVelocity = 1
)

// Start a fling given a starting velocity. Returns whether a
// fling was started.
func (f *Animation) Start(c unit.Metric, now time.Time, velocity float32) bool {
	min := float32(c.Px(minFlingVelocity))
	v := velocity
	if -min <= v && v <= min {
		return false
	}
	max := float32(c.Px(maxFlingVelocity))
	if v > max {
		v = max
	} else if v < -max {
		v = -max
	}
	f.init(now, v)
	return true
}

func (f *Animation) init(now time.Time, v0 float32) {
	f.t0 = now
	f.v0 = v0
	f.x = 0
}

func (f *Animation) Active() bool {
	return f.v0 != 0
}

// Tick computes and returns a fling distance since
// the last time Tick was called.
func (f *Animation) Tick(now time.Time) int {
	if !f.Active() {
		return 0
	}
	var k float32
	if runtime.GOOS == "darwin" {
		k = -2 // iOS
	} else {
		k = -4.2 // Android and default
	}
	t := now.Sub(f.t0)
	// The acceleration x''(t) of a point mass with a drag
	// force, f, proportional with velocity, x'(t), is
	// governed by the equation
	//
	// x''(t) = kx'(t)
	//
	// Given the starting position x(0) = 0, the starting
	// velocity x'(0) = v0, the position is then
	// given by
	//
	// x(t) = v0*e^(k*t)/k - v0/k
	//
	ekt := float32(math.Exp(float64(k) * t.Seconds()))
	x := f.v0*ekt/k - f.v0/k
	dist := x - f.x
	idist := int(dist)
	f.x += float32(idist)
	// Solving for the velocity x'(t) gives us
	//
	// x'(t) = v0*e^(k*t)
	v := f.v0 * ekt
	if -thresholdVelocity < v && v < thresholdVelocity {
		f.v0 = 0
	}
	return idist
}
