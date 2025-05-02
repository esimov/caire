// Go implementation of the StackBlur algorithm
// http://incubator.quasimondo.com/processing/fast_blur_deluxe.php

package caire

import (
	"errors"
	"image"
	"image/color"
)

// blurStack is a linked list containing the color value and a pointer to the next struct.
type blurStack struct {
	r, g, b, a uint32
	next       *blurStack
}

var mulTable = []uint32{
	512, 512, 456, 512, 328, 456, 335, 512, 405, 328, 271, 456, 388, 335, 292, 512,
	454, 405, 364, 328, 298, 271, 496, 456, 420, 388, 360, 335, 312, 292, 273, 512,
	482, 454, 428, 405, 383, 364, 345, 328, 312, 298, 284, 271, 259, 496, 475, 456,
	437, 420, 404, 388, 374, 360, 347, 335, 323, 312, 302, 292, 282, 273, 265, 512,
	497, 482, 468, 454, 441, 428, 417, 405, 394, 383, 373, 364, 354, 345, 337, 328,
	320, 312, 305, 298, 291, 284, 278, 271, 265, 259, 507, 496, 485, 475, 465, 456,
	446, 437, 428, 420, 412, 404, 396, 388, 381, 374, 367, 360, 354, 347, 341, 335,
	329, 323, 318, 312, 307, 302, 297, 292, 287, 282, 278, 273, 269, 265, 261, 512,
	505, 497, 489, 482, 475, 468, 461, 454, 447, 441, 435, 428, 422, 417, 411, 405,
	399, 394, 389, 383, 378, 373, 368, 364, 359, 354, 350, 345, 341, 337, 332, 328,
	324, 320, 316, 312, 309, 305, 301, 298, 294, 291, 287, 284, 281, 278, 274, 271,
	268, 265, 262, 259, 257, 507, 501, 496, 491, 485, 480, 475, 470, 465, 460, 456,
	451, 446, 442, 437, 433, 428, 424, 420, 416, 412, 408, 404, 400, 396, 392, 388,
	385, 381, 377, 374, 370, 367, 363, 360, 357, 354, 350, 347, 344, 341, 338, 335,
	332, 329, 326, 323, 320, 318, 315, 312, 310, 307, 304, 302, 299, 297, 294, 292,
	289, 287, 285, 282, 280, 278, 275, 273, 271, 269, 267, 265, 263, 261, 259,
}

var shgTable = []uint32{
	9, 11, 12, 13, 13, 14, 14, 15, 15, 15, 15, 16, 16, 16, 16, 17,
	17, 17, 17, 17, 17, 17, 18, 18, 18, 18, 18, 18, 18, 18, 18, 19,
	19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 19, 20, 20, 20,
	20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 21,
	21, 21, 21, 21, 21, 21, 21, 21, 21, 21, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22,
	22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 22, 23,
	23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23, 23,
	23, 23, 23, 23, 23, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
	24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24, 24,
}

// Stackblur takes the source image and returns it's blurred version by applying the blur radius defined as parameter. The destination image must be a image.NRGBA.
func Stackblur(dst, src image.Image, radius uint32) error {
	// Limit the maximum blur radius to 255 to avoid overflowing the multable.
	if int(radius) >= len(mulTable) {
		radius = uint32(len(mulTable) - 1)
	}

	if radius < 1 {
		return errors.New("blur radius must be greater than 0")
	}

	img, ok := dst.(*image.NRGBA)
	if !ok {
		return errors.New("the destination image must be image.NRGBA")
	}

	process(img, src, radius)
	return nil
}

func process(dst *image.NRGBA, src image.Image, radius uint32) {
	srcBounds := src.Bounds()
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y

	dstBounds := srcBounds.Sub(srcBounds.Min)
	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()

	switch src0 := src.(type) {
	case *image.NRGBA:
		rowSize := srcBounds.Dx() * 4
		for dstY := 0; dstY < dstH; dstY++ {
			di := src0.PixOffset(0, dstY)
			si := src0.PixOffset(srcMinX, srcMinY+dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				copy(dst.Pix[di:di+rowSize], src0.Pix[si:si+rowSize])
			}
		}
	case *image.YCbCr:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				srcX := srcMinX + dstX
				srcY := srcMinY + dstY
				siy := src0.YOffset(srcX, srcY)
				sic := src0.COffset(srcX, srcY)
				r, g, b := color.YCbCrToRGB(src0.Y[siy], src0.Cb[sic], src0.Cr[sic])
				dst.Pix[di+0] = r
				dst.Pix[di+1] = g
				dst.Pix[di+2] = b
				dst.Pix[di+3] = 0xff
				di += 4
			}
		}
	default:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				c := color.NRGBAModel.Convert(src.At(srcMinX+dstX, srcMinY+dstY)).(color.NRGBA)
				dst.Pix[di+0] = c.R
				dst.Pix[di+1] = c.G
				dst.Pix[di+2] = c.B
				dst.Pix[di+3] = c.A
				di += 4
			}
		}
	}

	blurImage(dst, radius)
}

func blurImage(src *image.NRGBA, radius uint32) {
	var (
		stackEnd *blurStack
		stackIn  *blurStack
		stackOut *blurStack
	)

	var width, height = uint32(src.Bounds().Dx()), uint32(src.Bounds().Dy())
	var (
		div, widthMinus1, heightMinus1, radiusPlus1, sumFactor uint32
		x, y, i, p, yp, yi, yw,
		rSum, gSum, bSum, aSum,
		rOutSum, gOutSum, bOutSum, aOutSum,
		rInSum, gInSum, bInSum, aInSum,
		pr, pg, pb, pa uint32
	)

	div = radius + radius + 1
	widthMinus1 = width - 1
	heightMinus1 = height - 1
	radiusPlus1 = radius + 1
	sumFactor = radiusPlus1 * (radiusPlus1 + 1) / 2

	stackStart := new(blurStack)
	stack := stackStart

	for i = 1; i < div; i++ {
		stack.next = new(blurStack)
		stack = stack.next
		if i == radiusPlus1 {
			stackEnd = stack
		}
	}
	stack.next = stackStart

	mulSum := mulTable[radius]
	shgSum := shgTable[radius]

	for y = 0; y < height; y++ {
		rInSum, gInSum, bInSum, aInSum, rSum, gSum, bSum, aSum = 0, 0, 0, 0, 0, 0, 0, 0

		pr = uint32(src.Pix[yi])
		pg = uint32(src.Pix[yi+1])
		pb = uint32(src.Pix[yi+2])
		pa = uint32(src.Pix[yi+3])

		rOutSum = radiusPlus1 * pr
		gOutSum = radiusPlus1 * pg
		bOutSum = radiusPlus1 * pb
		aOutSum = radiusPlus1 * pa

		rSum += sumFactor * pr
		gSum += sumFactor * pg
		bSum += sumFactor * pb
		aSum += sumFactor * pa

		stack = stackStart

		for i = 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa
			stack = stack.next
		}

		for i = 1; i < radiusPlus1; i++ {
			var diff uint32
			if widthMinus1 < i {
				diff = widthMinus1
			} else {
				diff = i
			}
			p = yi + (diff << 2)
			pr = uint32(src.Pix[p])
			pg = uint32(src.Pix[p+1])
			pb = uint32(src.Pix[p+2])
			pa = uint32(src.Pix[p+3])

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			rSum += stack.r * (radiusPlus1 - i)
			gSum += stack.g * (radiusPlus1 - i)
			bSum += stack.b * (radiusPlus1 - i)
			aSum += stack.a * (radiusPlus1 - i)

			rInSum += pr
			gInSum += pg
			bInSum += pb
			aInSum += pa

			stack = stack.next
		}
		stackIn = stackStart
		stackOut = stackEnd

		for x = 0; x < width; x++ {
			pa = (aSum * mulSum) >> shgSum
			src.Pix[yi+3] = uint8(pa)

			if pa != 0 {
				src.Pix[yi] = uint8((rSum * mulSum) >> shgSum)
				src.Pix[yi+1] = uint8((gSum * mulSum) >> shgSum)
				src.Pix[yi+2] = uint8((bSum * mulSum) >> shgSum)
			} else {
				src.Pix[yi] = 0
				src.Pix[yi+1] = 0
				src.Pix[yi+2] = 0
			}

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum
			aSum -= aOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b
			aOutSum -= stackIn.a

			p = x + radius + 1

			if p > widthMinus1 {
				p = widthMinus1
			}
			p = (yw + p) << 2

			stackIn.r = uint32(src.Pix[p])
			stackIn.g = uint32(src.Pix[p+1])
			stackIn.b = uint32(src.Pix[p+2])
			stackIn.a = uint32(src.Pix[p+3])

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b
			aInSum += stackIn.a

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum
			aSum += aInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb
			aOutSum += pa

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb
			aInSum -= pa

			stackOut = stackOut.next

			yi += 4
		}
		yw += width
	}

	for x = 0; x < width; x++ {
		rInSum, gInSum, bInSum, aInSum, rSum, gSum, bSum, aSum = 0, 0, 0, 0, 0, 0, 0, 0

		yi = x << 2
		pr = uint32(src.Pix[yi])
		pg = uint32(src.Pix[yi+1])
		pb = uint32(src.Pix[yi+2])
		pa = uint32(src.Pix[yi+3])

		rOutSum = radiusPlus1 * pr
		gOutSum = radiusPlus1 * pg
		bOutSum = radiusPlus1 * pb
		aOutSum = radiusPlus1 * pa

		rSum += sumFactor * pr
		gSum += sumFactor * pg
		bSum += sumFactor * pb
		aSum += sumFactor * pa

		stack = stackStart

		for i = 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa
			stack = stack.next
		}

		yp = width

		for i = 1; i <= radius; i++ {
			yi = (yp + x) << 2
			pr = uint32(src.Pix[yi])
			pg = uint32(src.Pix[yi+1])
			pb = uint32(src.Pix[yi+2])
			pa = uint32(src.Pix[yi+3])

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			rSum += stack.r * (radiusPlus1 - i)
			gSum += stack.g * (radiusPlus1 - i)
			bSum += stack.b * (radiusPlus1 - i)
			aSum += stack.a * (radiusPlus1 - i)

			rInSum += pr
			gInSum += pg
			bInSum += pb
			aInSum += pa

			stack = stack.next

			if i < heightMinus1 {
				yp += width
			}
		}

		yi = x
		stackIn = stackStart
		stackOut = stackEnd

		for y = 0; y < height; y++ {
			p = yi << 2
			pa = (aSum * mulSum) >> shgSum
			src.Pix[p+3] = uint8(pa)

			if pa > 0 {
				src.Pix[p] = uint8((rSum * mulSum) >> shgSum)
				src.Pix[p+1] = uint8((gSum * mulSum) >> shgSum)
				src.Pix[p+2] = uint8((bSum * mulSum) >> shgSum)
			} else {
				src.Pix[p] = 0
				src.Pix[p+1] = 0
				src.Pix[p+2] = 0
			}

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum
			aSum -= aOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b
			aOutSum -= stackIn.a

			p = y + radiusPlus1

			if p > heightMinus1 {
				p = heightMinus1
			}
			p = (x + (p * width)) << 2

			stackIn.r = uint32(src.Pix[p])
			stackIn.g = uint32(src.Pix[p+1])
			stackIn.b = uint32(src.Pix[p+2])
			stackIn.a = uint32(src.Pix[p+3])

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b
			aInSum += stackIn.a

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum
			aSum += aInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb
			aOutSum += pa

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb
			aInSum -= pa

			stackOut = stackOut.next

			yi += width
		}
	}
}
