package caire

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"os"
	"path/filepath"
	"testing"

	"github.com/esimov/caire/utils"
)

func TestImage_ShouldGetSampleImage(t *testing.T) {
	path := filepath.Join("./testdata", "sample.jpg")
	_, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Should get the sample image")
	}
}

func TestImage_ImgToNRGBA(t *testing.T) {
	rect := image.Rect(-1, -1, 15, 15)
	colors := palette.Plan9
	testCases := []struct {
		name string
		img  image.Image
	}{
		{
			name: "NRGBA",
			img:  makeNRGBAImage(rect, colors),
		},
		{
			name: "YCbCr-444",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio444),
		},
		{
			name: "YCbCr-422",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio422),
		},
		{
			name: "YCbCr-420",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio420),
		},
		{
			name: "YCbCr-440",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio440),
		},
		{
			name: "YCbCr-410",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio410),
		},
		{
			name: "YCbCr-411",
			img:  makeYCbCrImage(rect, colors, image.YCbCrSubsampleRatio411),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.img.Bounds()
			for y := r.Min.Y; y < r.Max.Y; y++ {
				buf := make([]byte, r.Dx()*4)
				scan(tc.img, 0, y-r.Min.Y, r.Dx(), y+1-r.Min.Y, buf)
				wantBuf := readRow(tc.img, y)
				if !compareBytes(buf, wantBuf, 1) {
					t.Errorf("scan horizontal line (y=%d): got %v want %v", y, buf, wantBuf)
				}
			}
			for x := r.Min.X; x < r.Max.X; x++ {
				buf := make([]byte, r.Dy()*4)
				scan(tc.img, x-r.Min.X, 0, x+1-r.Min.X, r.Dy(), buf)
				wantBuf := readColumn(tc.img, x)
				if !compareBytes(buf, wantBuf, 1) {
					t.Errorf("scan vertical line (x=%d): got %v want %v", x, buf, wantBuf)
				}
			}
		})
	}
}

func scan(img image.Image, x1, y1, x2, y2 int, dst []uint8) {
	switch img := img.(type) {
	case *image.NRGBA:
		size := (x2 - x1) * 4
		j := 0
		i := y1*img.Stride + x1*4
		for y := y1; y < y2; y++ {
			copy(dst[j:j+size], img.Pix[i:i+size])
			j += size
			i += img.Stride
		}
	case *image.YCbCr:
		j := 0
		x1 += img.Rect.Min.X
		x2 += img.Rect.Min.X
		y1 += img.Rect.Min.Y
		y2 += img.Rect.Min.Y
		for y := y1; y < y2; y++ {
			iy := (y-img.Rect.Min.Y)*img.YStride + (x1 - img.Rect.Min.X)
			for x := x1; x < x2; x++ {
				var ic int
				switch img.SubsampleRatio {
				case image.YCbCrSubsampleRatio444:
					ic = (y-img.Rect.Min.Y)*img.CStride + (x - img.Rect.Min.X)
				case image.YCbCrSubsampleRatio422:
					ic = (y-img.Rect.Min.Y)*img.CStride + (x/2 - img.Rect.Min.X/2)
				case image.YCbCrSubsampleRatio420:
					ic = (y/2-img.Rect.Min.Y/2)*img.CStride + (x/2 - img.Rect.Min.X/2)
				case image.YCbCrSubsampleRatio440:
					ic = (y/2-img.Rect.Min.Y/2)*img.CStride + (x - img.Rect.Min.X)
				default:
					ic = img.COffset(x, y)
				}

				yy := int(img.Y[iy])
				cb := int(img.Cb[ic]) - 128
				cr := int(img.Cr[ic]) - 128

				r := (yy<<16 + 91881*cr + 1<<15) >> 16
				if r > 0xff {
					r = 0xff
				} else if r < 0 {
					r = 0
				}

				g := (yy<<16 - 22554*cb - 46802*cr + 1<<15) >> 16
				if g > 0xff {
					g = 0xff
				} else if g < 0 {
					g = 0
				}

				b := (yy<<16 + 116130*cb + 1<<15) >> 16
				if b > 0xff {
					b = 0xff
				} else if b < 0 {
					b = 0
				}

				dst[j+0] = uint8(r)
				dst[j+1] = uint8(g)
				dst[j+2] = uint8(b)
				dst[j+3] = 0xff

				iy++
				j += 4
			}
		}
	}
}

func makeYCbCrImage(rect image.Rectangle, colors []color.Color, sr image.YCbCrSubsampleRatio) *image.YCbCr {
	img := image.NewYCbCr(rect, sr)
	j := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			iy := img.YOffset(x, y)
			ic := img.COffset(x, y)
			c := color.NRGBAModel.Convert(colors[j]).(color.NRGBA)
			img.Y[iy], img.Cb[ic], img.Cr[ic] = color.RGBToYCbCr(c.R, c.G, c.B)
			j++
		}
	}
	return img
}

func makeNRGBAImage(rect image.Rectangle, colors []color.Color) *image.NRGBA {
	img := image.NewNRGBA(rect)
	fillDrawImage(img, colors)
	return img
}

func fillDrawImage(img draw.Image, colors []color.Color) {
	colorsNRGBA := make([]color.NRGBA, len(colors))
	for i, c := range colors {
		nrgba := color.NRGBAModel.Convert(c).(color.NRGBA)
		nrgba.A = uint8(i % 256)
		colorsNRGBA[i] = nrgba
	}
	rect := img.Bounds()
	i := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, colorsNRGBA[i])
			i++
		}
	}
}

func readRow(img image.Image, y int) []uint8 {
	row := make([]byte, img.Bounds().Dx()*4)
	i := 0
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
		row[i+0] = c.R
		row[i+1] = c.G
		row[i+2] = c.B
		row[i+3] = c.A
		i += 4
	}
	return row
}

func readColumn(img image.Image, x int) []uint8 {
	column := make([]byte, img.Bounds().Dy()*4)
	i := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
		column[i+0] = c.R
		column[i+1] = c.G
		column[i+2] = c.B
		column[i+3] = c.A
		i += 4
	}
	return column
}

func compareBytes(a, b []uint8, delta int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if utils.Abs(int(a[i])-int(b[i])) > delta {
			return false
		}
	}
	return true
}
