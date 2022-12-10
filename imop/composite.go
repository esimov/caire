// Package imop implements the Porter-Duff composition operations
// used for mixing a graphic element with its backdrop.
// Porter and Duff presented in their paper 12 different composition operation, but the
// core image/draw core package implements only the source-over-destination and source.
// This package implements all of the 12 composite operation together with some blending modes.
package imop

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/esimov/caire/utils"
)

const (
	Clear   = "clear"
	Copy    = "copy"
	Dst     = "dst"
	SrcOver = "src_over"
	DstOver = "dst_over"
	SrcIn   = "src_in"
	DstIn   = "dst_in"
	SrcOut  = "src_out"
	DstOut  = "dst_out"
	SrcAtop = "src_atop"
	DstAtop = "dst_atop"
	Xor     = "xor"
)

// Bitmap holds an image type as a placeholder for the Porter-Duff composition
// operations which can be used as a source or destination image.
type Bitmap struct {
	Img *image.NRGBA
}

// Composite struct contains the currently active composition operation and all the supported operations.
type Composite struct {
	CurrentOp string
	Ops       []string
}

// NewBitmap initializes a new Bitmap.
func NewBitmap(rect image.Rectangle) *Bitmap {
	return &Bitmap{
		Img: image.NewNRGBA(rect),
	}
}

// InitOp initializes a new composition operation.
func InitOp() *Composite {
	return &Composite{
		CurrentOp: SrcOver,
		Ops: []string{
			Clear,
			Copy,
			Dst,
			SrcOver,
			DstOver,
			SrcIn,
			DstIn,
			SrcOut,
			DstOut,
			SrcAtop,
			DstAtop,
			Xor,
		},
	}
}

// Set changes the current composition operation.
func (op *Composite) Set(cop string) error {
	if utils.Contains(op.Ops, cop) {
		op.CurrentOp = cop
		return nil
	}
	return fmt.Errorf("unsupported composition operation")
}

// Set changes the current composition operation.
func (op *Composite) Get() string {
	return op.CurrentOp
}

// Draw applies the currently active Ported-Duff composition operation formula,
// taking as parameter the source and the destination image and draws the result into the bitmap.
// If a blend mode is activated it will plug in the alpha blending formula also into the equation.
func (op *Composite) Draw(bitmap *Bitmap, src, dst *image.NRGBA, bl *Blend) {
	dx, dy := src.Bounds().Dx(), src.Bounds().Dy()

	var (
		r, g, b, a     uint32
		rn, gn, bn, an float64
	)

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r1, g1, b1, a1 := src.At(x, y).RGBA()
			r2, g2, b2, a2 := dst.At(x, y).RGBA()

			rs, gs, bs, as := r1>>8, g1>>8, b1>>8, a1>>8
			rb, gb, bb, ab := r2>>8, g2>>8, b2>>8, a2>>8

			// normalize the values.
			rsn := float64(rs) / 255
			gsn := float64(gs) / 255
			bsn := float64(bs) / 255
			asn := float64(as) / 255

			rbn := float64(rb) / 255
			gbn := float64(gb) / 255
			bbn := float64(bb) / 255
			abn := float64(ab) / 255

			// applying the alpha composition formula
			switch op.CurrentOp {
			case Clear:
				rn, gn, bn, an = 0, 0, 0, 0
			case Copy:
				rn = asn * rsn
				gn = asn * gsn
				bn = asn * bsn
				an = asn * asn
			case Dst:
				rn = abn * rbn
				gn = abn * gbn
				bn = abn * bbn
				an = abn * abn
			case SrcOver:
				rn = asn*rsn + abn*rbn*(1-asn)
				gn = asn*gsn + abn*gbn*(1-asn)
				bn = asn*bsn + abn*bbn*(1-asn)
				an = asn + abn*(1-asn)
			case DstOver:
				rn = asn*rsn*(1-abn) + abn*rbn
				gn = asn*gsn*(1-abn) + abn*gbn
				bn = asn*bsn*(1-abn) + abn*bbn
				an = asn*(1-abn) + abn
			case SrcIn:
				rn = asn * rsn * abn
				gn = asn * gsn * abn
				bn = asn * bsn * abn
				an = asn * abn
			case DstIn:
				rn = abn * rbn * asn
				gn = abn * gbn * asn
				bn = abn * bbn * asn
				an = abn * asn
			case SrcOut:
				rn = asn * rsn * (1 - abn)
				gn = asn * gsn * (1 - abn)
				bn = asn * bsn * (1 - abn)
				an = asn * (1 - abn)
			case DstOut:
				rn = abn * rbn * (1 - asn)
				gn = abn * gbn * (1 - asn)
				bn = abn * bbn * (1 - asn)
				an = abn * (1 - asn)
			case SrcAtop:
				rn = asn*rsn*abn + (1-asn)*abn*rbn
				gn = asn*gsn*abn + (1-asn)*abn*gbn
				bn = asn*bsn*abn + (1-asn)*abn*bbn
				an = asn*abn + abn*(1-asn)
			case DstAtop:
				rn = asn*rsn*(1-abn) + abn*rbn*asn
				gn = asn*gsn*(1-abn) + abn*gbn*asn
				bn = asn*bsn*(1-abn) + abn*bbn*asn
				an = asn*(1-abn) + abn*asn
			case Xor:
				rn = asn*rsn*(1-abn) + abn*rbn*(1-asn)
				gn = asn*gsn*(1-abn) + abn*gbn*(1-asn)
				bn = asn*bsn*(1-abn) + abn*bbn*(1-asn)
				an = asn*(1-abn) + abn*(1-asn)
			}

			r = uint32(rn * 255)
			g = uint32(gn * 255)
			b = uint32(bn * 255)
			a = uint32(an * 255)

			bitmap.Img.Set(x, y, color.NRGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			})

			// applying the blending mode
			if bl != nil {
				rn, gn, bn, an = 0, 0, 0, 0 // reset the colors
				r1, g1, b1, a1 = src.At(x, y).RGBA()
				r2, g2, b2, a2 = dst.At(x, y).RGBA()

				rs, gs, bs, as = r1>>8, g1>>8, b1>>8, a1>>8
				rb, gb, bb, ab = r2>>8, g2>>8, b2>>8, a2>>8

				rsn = float64(rs) / 255
				gsn = float64(gs) / 255
				bsn = float64(bs) / 255
				asn = float64(as) / 255

				rbn = float64(rb) / 255
				gbn = float64(gb) / 255
				bbn = float64(bb) / 255
				abn = float64(ab) / 255

				foreground := Color{R: rsn, G: gsn, B: bsn}
				background := Color{R: rbn, G: gbn, B: bbn}

				switch bl.Current {
				case Normal:
					rn, gn, bn, an = rsn, gsn, bsn, asn
				case Darken:
					rn = utils.Min(rsn, rbn)
					gn = utils.Min(gsn, gbn)
					bn = utils.Min(bsn, bbn)
					an = utils.Min(asn, abn)
				case Lighten:
					rn = utils.Max(rsn, rbn)
					gn = utils.Max(gsn, gbn)
					bn = utils.Max(bsn, bbn)
					an = utils.Max(asn, abn)
				case Screen:
					rn = 1 - (1-rsn)*(1-rbn)
					gn = 1 - (1-gsn)*(1-gbn)
					bn = 1 - (1-bsn)*(1-bbn)
					an = 1 - (1-asn)*(1-abn)
				case Multiply:
					rn = rsn * rbn
					gn = gsn * gbn
					bn = bsn * bbn
					an = asn * abn
				case Overlay:
					if rsn <= 0.5 {
						rn = 2 * rsn * rbn
					} else {
						rn = 1 - 2*(1-rsn)*(1-rbn)
					}

					if gsn <= 0.5 {
						gn = 2 * gsn * gbn
					} else {
						gn = 1 - 2*(1-gsn)*(1-gbn)
					}

					if bsn <= 0.5 {
						bn = 2 * bsn * bbn
					} else {
						bn = 1 - 2*(1-bsn)*(1-bbn)
					}

					if asn <= 0.5 {
						an = 2 * asn * abn
					} else {
						an = 1 - 2*(1-asn)*(1-abn)
					}
				case SoftLight:
					if rbn < 0.5 {
						rn = rsn - (1-2*rbn)*rsn*(1-rsn)
					} else {
						var w3r float64
						if rsn < 0.25 {
							w3r = ((16*rsn-12)*rsn + 4) * rsn
						} else {
							w3r = math.Sqrt(rsn)
						}
						rn = rsn + (2*rbn-1)*(w3r-rsn)
					}

					if gbn < 0.5 {
						gn = gsn - (1-2*gbn)*gsn*(1-gsn)
					} else {
						var w3g float64
						if gsn < 0.25 {
							w3g = ((16*gsn-12)*gsn + 4) * gsn
						} else {
							w3g = math.Sqrt(gsn)
						}
						gn = gsn + (2*gbn-1)*(w3g-gsn)
					}

					if bbn < 0.5 {
						bn = bsn - (1-2*bbn)*bsn*(1-bsn)
					} else {
						var w3b float64
						if bsn < 0.25 {
							w3b = ((16*bsn-12)*bsn + 4) * bsn
						} else {
							w3b = math.Sqrt(bsn)
						}
						bn = bsn + (2*bbn-1)*(w3b-bsn)
					}

					if abn < 0.5 {
						an = asn - (1-2*abn)*asn*(1-asn)
					} else {
						var w3a float64
						if asn < 0.25 {
							w3a = ((16*asn-12)*asn + 4) * asn
						} else {
							w3a = math.Sqrt(asn)
						}
						an = asn + (2*abn-1)*(w3a-asn)
					}
				case HardLight:
					if rbn < 0.5 {
						rn = rbn - (1-2*rsn)*rbn*(1-rbn)
					} else {
						var w3r float64
						if rbn < 0.25 {
							w3r = ((16*rbn-12)*rbn + 4) * rbn
						} else {
							w3r = math.Sqrt(rbn)
						}
						rn = rbn + (2*rsn-1)*(w3r-rbn)
					}

					if gbn < 0.5 {
						gn = gbn - (1-2*gsn)*gbn*(1-gbn)
					} else {
						var w3g float64
						if gbn < 0.25 {
							w3g = ((16*gbn-12)*gbn + 4) * gbn
						} else {
							w3g = math.Sqrt(gbn)
						}
						gn = gbn + (2*gsn-1)*(w3g-gbn)
					}

					if bbn < 0.5 {
						bn = bbn - (1-2*bsn)*bbn*(1-bbn)
					} else {
						var w3b float64
						if bbn < 0.25 {
							w3b = ((16*bbn-12)*bbn + 4) * bbn
						} else {
							w3b = math.Sqrt(bbn)
						}
						bn = bbn + (2*bsn-1)*(w3b-bbn)
					}

					if abn < 0.5 {
						an = abn - (1-2*asn)*abn*(1-abn)
					} else {
						var w3a float64
						if abn < 0.25 {
							w3a = ((16*abn-12)*abn + 4) * abn
						} else {
							w3a = math.Sqrt(abn)
						}
						an = abn + (2*asn-1)*(w3a-abn)
					}
				case ColorDodge:
					if rsn < 1 {
						rn = utils.Min(1, rbn/(1-rsn))
					} else if rsn == 1 {
						rn = 1
					}

					if gsn < 1 {
						gn = utils.Min(1, gbn/(1-gsn))
					} else if gsn == 1 {
						gn = 1
					}

					if bsn < 1 {
						bn = utils.Min(1, bbn/(1-bsn))
					} else if bsn == 1 {
						bn = 1
					}

					if asn < 1 {
						an = utils.Min(1, abn/(1-asn))
					} else if asn == 1 {
						an = 1
					}
				case ColorBurn:
					if rsn > 0 {
						rn = 1 - utils.Min(1, (1-rbn)/rsn)
					} else if rsn == 0 {
						rn = 0
					}

					if gsn > 0 {
						gn = 1 - utils.Min(1, (1-gbn)/gsn)
					} else if gsn == 0 {
						gn = 0
					}

					if bsn > 0 {
						bn = 1 - utils.Min(1, (1-bbn)/bsn)
					} else if bsn == 0 {
						bn = 0
					}

					if asn > 0 {
						an = 1 - utils.Min(1, (1-abn)/asn)
					} else if asn == 0 {
						an = 0
					}
				case Difference:
					rn = utils.Abs(rbn - rsn)
					gn = utils.Abs(gbn - gsn)
					bn = utils.Abs(bbn - bsn)
					an = 1
				case Exclusion:
					rn = rsn + rbn - 2*rsn*rbn
					gn = gsn + gbn - 2*gsn*gbn
					bn = bsn + bbn - 2*bsn*bbn
					an = 1

				// Non-separable blend modes
				// https://www.w3.org/TR/compositing-1/#blendingnonseparable
				case Hue:
					sat := bl.SetSat(background, bl.Sat(foreground))
					rgb := bl.SetLum(sat, bl.Lum(foreground))

					a := asn + abn - asn*abn
					rn = bl.AlphaCompose(abn, asn, a, rbn*255, rsn*255, rgb.R*255)
					gn = bl.AlphaCompose(abn, asn, a, gbn*255, gsn*255, rgb.G*255)
					bn = bl.AlphaCompose(abn, asn, a, bbn*255, bsn*255, rgb.B*255)
					rn, gn, bn = rn/255, gn/255, bn/255
					an = a
				case Saturation:
					sat := bl.SetSat(foreground, bl.Sat(background))
					rgb := bl.SetLum(sat, bl.Lum(foreground))

					a := asn + abn - asn*abn
					rn = bl.AlphaCompose(abn, asn, a, rbn*255, rsn*255, rgb.R*255)
					gn = bl.AlphaCompose(abn, asn, a, gbn*255, gsn*255, rgb.G*255)
					bn = bl.AlphaCompose(abn, asn, a, bbn*255, bsn*255, rgb.B*255)
					rn, gn, bn = rn/255, gn/255, bn/255
					an = a
				case ColorMode:
					rgb := bl.SetLum(background, bl.Lum(foreground))

					a := asn + abn - asn*abn
					rn = bl.AlphaCompose(abn, asn, a, rbn*255, rsn*255, rgb.R*255)
					gn = bl.AlphaCompose(abn, asn, a, gbn*255, gsn*255, rgb.G*255)
					bn = bl.AlphaCompose(abn, asn, a, bbn*255, bsn*255, rgb.B*255)
					rn, gn, bn = rn/255, gn/255, bn/255
					an = a
				case Luminosity:
					rgb := bl.SetLum(foreground, bl.Lum(background))

					a := asn + abn - asn*abn
					rn = bl.AlphaCompose(abn, asn, a, rbn*255, rsn*255, rgb.R*255)
					gn = bl.AlphaCompose(abn, asn, a, gbn*255, gsn*255, rgb.G*255)
					bn = bl.AlphaCompose(abn, asn, a, bbn*255, bsn*255, rgb.B*255)
					rn, gn, bn = rn/255, gn/255, bn/255
					an = a
				}
			}

			r = uint32(rn * 255)
			g = uint32(gn * 255)
			b = uint32(bn * 255)
			a = uint32(an * 255)

			bitmap.Img.Set(x, y, color.NRGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			})
		}
	}
}
