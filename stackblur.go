// Go implementation of StackBlur algorithm described here:
// http://incubator.quasimondo.com/processing/fast_blur_deluxe.php

package caire

import (
	"image"
)

type blurstack struct {
	r, g, b, a uint32
	next       *blurstack
}

var mul_table []uint32 = []uint32{
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

var shg_table []uint32 = []uint32{
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

func (bs *blurstack) NewBlurStack() *blurstack {
	return &blurstack{bs.r, bs.g, bs.b, bs.a, bs.next}
}

func Stackblur(img *image.NRGBA, width, height, radius uint32) *image.NRGBA {
	var stackEnd, stackIn, stackOut *blurstack
	var (
		div, widthMinus1, heightMinus1, radiusPlus1, sumFactor uint32
		x, y, i, p, yp, yi, yw,
		r_sum, g_sum, b_sum, a_sum,
		r_out_sum, g_out_sum, b_out_sum, a_out_sum,
		r_in_sum, g_in_sum, b_in_sum, a_in_sum,
		pr, pg, pb, pa uint32
	)

	div = radius + radius + 1
	widthMinus1 = width - 1
	heightMinus1 = height - 1
	radiusPlus1 = radius + 1
	sumFactor = radiusPlus1 * (radiusPlus1 + 1) / 2

	bs := blurstack{}
	stackStart := bs.NewBlurStack()
	stack := stackStart

	for i = 1; i < div; i++ {
		stack.next = bs.NewBlurStack()
		stack = stack.next
		if i == radiusPlus1 {
			stackEnd = stack
		}
	}
	stack.next = stackStart

	mul_sum := mul_table[radius]
	shg_sum := shg_table[radius]

	for y = 0; y < height; y++ {
		r_in_sum, g_in_sum, b_in_sum, a_in_sum, r_sum, g_sum, b_sum, a_sum = 0, 0, 0, 0, 0, 0, 0, 0

		pr = uint32(img.Pix[yi])
		pg = uint32(img.Pix[yi+1])
		pb = uint32(img.Pix[yi+2])
		pa = uint32(img.Pix[yi+3])

		r_out_sum = radiusPlus1 * pr
		g_out_sum = radiusPlus1 * pg
		b_out_sum = radiusPlus1 * pb
		a_out_sum = radiusPlus1 * pa

		r_sum += sumFactor * pr
		g_sum += sumFactor * pg
		b_sum += sumFactor * pb
		a_sum += sumFactor * pa

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
			pr = uint32(img.Pix[p])
			pg = uint32(img.Pix[p+1])
			pb = uint32(img.Pix[p+2])
			pa = uint32(img.Pix[p+3])

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			r_sum += stack.r * (radiusPlus1 - i)
			g_sum += stack.g * (radiusPlus1 - i)
			b_sum += stack.b * (radiusPlus1 - i)
			a_sum += stack.a * (radiusPlus1 - i)

			r_in_sum += pr
			g_in_sum += pg
			b_in_sum += pb
			a_in_sum += pa

			stack = stack.next
		}
		stackIn = stackStart
		stackOut = stackEnd

		for x = 0; x < width; x++ {
			pa = (a_sum * mul_sum) >> shg_sum
			img.Pix[yi+3] = uint8(pa)

			if pa != 0 {
				pa = 255 / pa
				img.Pix[yi] = uint8((r_sum * mul_sum) >> shg_sum)
				img.Pix[yi+1] = uint8((g_sum * mul_sum) >> shg_sum)
				img.Pix[yi+2] = uint8((b_sum * mul_sum) >> shg_sum)
			} else {
				img.Pix[yi] = 0
				img.Pix[yi+1] = 0
				img.Pix[yi+2] = 0
			}

			r_sum -= r_out_sum
			g_sum -= g_out_sum
			b_sum -= b_out_sum
			a_sum -= a_out_sum

			r_out_sum -= stackIn.r
			g_out_sum -= stackIn.g
			b_out_sum -= stackIn.b
			a_out_sum -= stackIn.a

			p = x + radius + 1

			if p > widthMinus1 {
				p = widthMinus1
			}
			p = (yw + p) << 2

			stackIn.r = uint32(img.Pix[p])
			stackIn.g = uint32(img.Pix[p+1])
			stackIn.b = uint32(img.Pix[p+2])
			stackIn.a = uint32(img.Pix[p+3])

			r_in_sum += stackIn.r
			g_in_sum += stackIn.g
			b_in_sum += stackIn.b
			a_in_sum += stackIn.a

			r_sum += r_in_sum
			g_sum += g_in_sum
			b_sum += b_in_sum
			a_sum += a_in_sum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			r_out_sum += pr
			g_out_sum += pg
			b_out_sum += pb
			a_out_sum += pa

			r_in_sum -= pr
			g_in_sum -= pg
			b_in_sum -= pb
			a_in_sum -= pa

			stackOut = stackOut.next

			yi += 4
		}
		yw += width
	}

	for x = 0; x < width; x++ {
		r_in_sum, g_in_sum, b_in_sum, a_in_sum, r_sum, g_sum, b_sum, a_sum = 0, 0, 0, 0, 0, 0, 0, 0

		yi = x << 2
		pr = uint32(img.Pix[yi])
		pg = uint32(img.Pix[yi+1])
		pb = uint32(img.Pix[yi+2])
		pa = uint32(img.Pix[yi+3])

		r_out_sum = radiusPlus1 * pr
		g_out_sum = radiusPlus1 * pg
		b_out_sum = radiusPlus1 * pb
		a_out_sum = radiusPlus1 * pa

		r_sum += sumFactor * pr
		g_sum += sumFactor * pg
		b_sum += sumFactor * pb
		a_sum += sumFactor * pa

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
			pr = uint32(img.Pix[yi])
			pg = uint32(img.Pix[yi+1])
			pb = uint32(img.Pix[yi+2])
			pa = uint32(img.Pix[yi+3])

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			r_sum += stack.r * (radiusPlus1 - i)
			g_sum += stack.g * (radiusPlus1 - i)
			b_sum += stack.b * (radiusPlus1 - i)
			a_sum += stack.a * (radiusPlus1 - i)

			r_in_sum += pr
			g_in_sum += pg
			b_in_sum += pb
			a_in_sum += pa

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
			pa = (a_sum * mul_sum) >> shg_sum
			img.Pix[p+3] = uint8(pa)

			if pa > 0 {
				pa = 255 / pa
				img.Pix[p] = uint8((r_sum * mul_sum) >> shg_sum)
				img.Pix[p+1] = uint8((g_sum * mul_sum) >> shg_sum)
				img.Pix[p+2] = uint8((b_sum * mul_sum) >> shg_sum)
			} else {
				img.Pix[p] = 0
				img.Pix[p+1] = 0
				img.Pix[p+2] = 0
			}

			r_sum -= r_out_sum
			g_sum -= g_out_sum
			b_sum -= b_out_sum
			a_sum -= a_out_sum

			r_out_sum -= stackIn.r
			g_out_sum -= stackIn.g
			b_out_sum -= stackIn.b
			a_out_sum -= stackIn.a

			p = y + radiusPlus1

			if p > heightMinus1 {
				p = heightMinus1
			}
			p = (x + (p * width)) << 2

			stackIn.r = uint32(img.Pix[p])
			stackIn.g = uint32(img.Pix[p+1])
			stackIn.b = uint32(img.Pix[p+2])
			stackIn.a = uint32(img.Pix[p+3])

			r_in_sum += stackIn.r
			g_in_sum += stackIn.g
			b_in_sum += stackIn.b
			a_in_sum += stackIn.a

			r_sum += r_in_sum
			g_sum += g_in_sum
			b_sum += b_in_sum
			a_sum += a_in_sum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			r_out_sum += pr
			g_out_sum += pg
			b_out_sum += pb
			a_out_sum += pa

			r_in_sum -= pr
			g_in_sum -= pg
			b_in_sum -= pb
			a_in_sum -= pa

			stackOut = stackOut.next

			yi += width
		}
	}
	return img
}
