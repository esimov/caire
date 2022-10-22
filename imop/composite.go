// Package imop implements the Porter-Duff composition operations
// used for mixing a graphic element with its backdrop.
// Porter and Duff presented in their paper 12 different composition operation,
// but the image/draw core package implements only the source-over-destination and source.
// This package is aimed to overcome the missing composite operations.

// It is mainly used to debug the seam carving operation correctness
// with face detection and image mask enabled.
// When the GUI mode and the debugging option is activated it will show
// the image mask and the detected faces rectangles in a distinct color.
package imop

import (
	"image"
	"image/color"

	"github.com/esimov/caire/utils"
)

const (
	Copy    = "copy"
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

// Composite defines a struct with the active and all the supported composition operations.
type Composite struct {
	currentOp string
	ops       []string
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
		currentOp: Copy,
		ops: []string{
			Copy,
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
func (op *Composite) Set(cop string) {
	op.currentOp = cop
}

// Draw applies the currently active Ported-Duff composition operation formula,
// taking as parameter the source and the destination image and draws the result into the bitmap.
// If a blend mode is activated it will plug in the alpha blending formula also into the equation.
func (op *Composite) Draw(bitmap *Bitmap, src, dst *image.NRGBA, blend *Blend) {
	dx, dy := src.Bounds().Dx(), src.Bounds().Dy()
	if bitmap == nil {
		bitmap = NewBitmap(src.Bounds())
	}

	var (
		r, g, b, a     uint32
		rn, gn, bn, an float64
	)

	if utils.Contains(op.ops, op.currentOp) {
		for x := 0; x < dx; x++ {
			for y := 0; y < dy; y++ {
				r1, g1, b1, a1 := src.At(x, y).RGBA()
				r2, g2, b2, a2 := dst.At(x, y).RGBA()

				rs, gs, bs, as := r1>>8, g1>>8, b1>>8, a1>>8
				rb, gb, bb, ab := r2>>8, g2>>8, b2>>8, a2>>8

				rsn := float64(rs) / 255
				gsn := float64(gs) / 255
				bsn := float64(bs) / 255
				asn := float64(as) / 255

				rbn := float64(rb) / 255
				gbn := float64(gb) / 255
				bbn := float64(bb) / 255
				abn := float64(ab) / 255

				// applying the alpha composition formula
				switch op.currentOp {
				case Copy:
					rn = asn * rsn
					gn = asn * gsn
					bn = asn * bsn
					an = 1
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
				if blend != nil {
					r1, g1, b1, a1 = bitmap.Img.At(x, y).RGBA()
					r2, g2, b2, a2 = src.At(x, y).RGBA()

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

					switch blend.OpType {
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
}
