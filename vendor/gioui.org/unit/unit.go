// SPDX-License-Identifier: Unlicense OR MIT

/*
Package unit implements device independent units.

Device independent pixel, or dp, is the unit for sizes independent of
the underlying display device.

Scaled pixels, or sp, is the unit for text sizes. An sp is like dp with
text scaling applied.

Finally, pixels, or px, is the unit for display dependent pixels. Their
size vary between platforms and displays.

To maintain a constant visual size across platforms and displays, always
use dps or sps to define user interfaces. Only use pixels for derived
values.
*/
package unit

import (
	"math"
)

// Metric converts Values to device-dependent pixels, px. The zero
// value represents a 1-to-1 scale from dp, sp to pixels.
type Metric struct {
	// PxPerDp is the device-dependent pixels per dp.
	PxPerDp float32
	// PxPerSp is the device-dependent pixels per sp.
	PxPerSp float32
}

type (
	// Dp represents device independent pixels. 1 dp will
	// have the same apparent size across platforms and
	// display resolutions.
	Dp float32
	// Sp is like UnitDp but for font sizes.
	Sp float32
)

// Dp converts v to pixels, rounded to the nearest integer value.
func (c Metric) Dp(v Dp) int {
	return int(math.Round(float64(nonZero(c.PxPerDp)) * float64(v)))
}

// Sp converts v to pixels, rounded to the nearest integer value.
func (c Metric) Sp(v Sp) int {
	return int(math.Round(float64(nonZero(c.PxPerSp)) * float64(v)))
}

// DpToSp converts v dp to sp.
func (c Metric) DpToSp(v Dp) Sp {
	return Sp(float32(v) * nonZero(c.PxPerDp) / nonZero(c.PxPerSp))
}

// SpToDp converts v sp to dp.
func (c Metric) SpToDp(v Sp) Dp {
	return Dp(float32(v) * nonZero(c.PxPerSp) / nonZero(c.PxPerDp))
}

// PxToSp converts v px to sp.
func (c Metric) PxToSp(v int) Sp {
	return Sp(float32(v) / nonZero(c.PxPerSp))
}

// PxToDp converts v px to dp.
func (c Metric) PxToDp(v int) Dp {
	return Dp(float32(v) / nonZero(c.PxPerDp))
}

func nonZero(v float32) float32 {
	if v == 0. {
		return 1
	}
	return v
}
