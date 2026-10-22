package caire

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/esimov/caire/utils"
	"github.com/stretchr/testify/assert"
)

func TestResize_ShrinkImageWidth(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newWidth := imgWidth / 2

	p.NewWidth = newWidth
	p.NewHeight = imgHeight

	for x := 0; x < newWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		seams := c.FindLowestEnergySeams(p)
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	imgWidth := img.Bounds().Max.X

	assert.Equal(t, imgWidth, newWidth)
}

func TestResize_ShrinkImageHeight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := imgHeight / 2

	p.NewWidth = imgWidth
	p.NewHeight = newHeight

	img = rotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		seams := c.FindLowestEnergySeams(p)
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	img = rotateImage270(img)
	imgHeight := img.Bounds().Max.Y

	assert.Equal(t, imgHeight, newHeight)
}

func TestResize_EnlargeImageWidth(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newWidth := imgWidth * 2

	p.NewWidth = newWidth
	p.NewHeight = imgHeight

	for x := 0; x < newWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		seams := c.FindLowestEnergySeams(p)
		img = c.AddSeam(img, seams, p.Debug)
	}
	imgWidth := img.Bounds().Max.X - img.Bounds().Dx()

	assert.NotEqual(t, imgWidth, newWidth)
}

func TestResize_EnlargeImageHeight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := imgHeight * 2

	p.NewWidth = imgWidth
	p.NewHeight = newHeight

	img = rotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		seams := c.FindLowestEnergySeams(p)
		img = c.AddSeam(img, seams, p.Debug)
	}
	img = rotateImage270(img)
	imgHeight := img.Bounds().Max.Y - img.Bounds().Dy()

	assert.NotEqual(t, imgHeight, newHeight)
}

// squareTestImage returns a 60x40 image with a uniform black square on a
// triangle wave background. The wave has a slope of 8 per pixel along x and
// an even energy of about 64, except for the peak column at x=32, which the
// square covers. The square has no energy inside, so the seams run through
// it unless a mask protects it.
func squareTestImage(square image.Rectangle, fill color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 60, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 60; x++ {
			v := uint8(utils.Abs(32-x) * 8)
			img.SetNRGBA(x, y, color.NRGBA{R: 255 - v, G: 255 - v, B: 255 - v, A: 0xff})
		}
	}
	draw.Draw(img, square, &image.Uniform{fill}, image.Point{}, draw.Src)
	return img
}

// resizeThroughProcess runs the Processor over the image the way the CLI does
// and decodes the result. The PNG extension selects a lossless encoding.
func resizeThroughProcess(t *testing.T, proc *Processor, img image.Image) image.Image {
	var in bytes.Buffer
	assert.NoError(t, png.Encode(&in, img))
	outPath := filepath.Join(t.TempDir(), "out.png")
	out, err := os.Create(outPath)
	assert.NoError(t, err)
	assert.NoError(t, proc.Process(&in, out))
	assert.NoError(t, out.Close())

	f, err := os.Open(outPath)
	assert.NoError(t, err)
	defer f.Close()
	res, _, err := image.Decode(f)
	assert.NoError(t, err)
	return res
}

// countColor returns the number of pixels of exactly the given color.
func countColor(img image.Image, want color.NRGBA) int {
	count := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) == want {
				count++
			}
		}
	}
	return count
}

func TestResize_MasksShouldFollowTheImageOnEveryResizePath(t *testing.T) {
	square := image.Rect(20, 10, 40, 30)
	black := color.NRGBA{A: 0xff}

	// A protective mask two pixels larger than the square keeps all of its
	// 400 pixels, whichever way the image is resized. A removal mask over the
	// full width of the square's rows attracts every horizontal seam, so a
	// height shrink by 10 takes exactly 10 of the 20 square rows away.
	// The vertical passes rotate the image. The masks used to stay unrotated,
	// so they no longer covered the square.
	// The order matters on versions where the both-axis flag and the table
	// of added seams are package variables: the flag stays set after the
	// first both-axis case, and the table keeps the seams of every
	// enlargement. The shrink-only cases run first and the mixed case next.
	cases := []struct {
		name                string
		newWidth, newHeight int
		region              image.Rectangle
		removal             bool
		wantSquarePixels    int
	}{
		{"shrink height only", 0, 30, square.Inset(-2), false, 400},
		{"shrink height only with a removal mask", 0, 30, image.Rect(0, 10, 60, 30), true, 200},
		{"enlarge width and shrink height", 70, 30, square.Inset(-2), false, 400},
		{"shrink width and enlarge height", 50, 45, square.Inset(-2), false, 400},
		{"enlarge height only", 0, 45, square.Inset(-2), false, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := squareTestImage(square, black)
			mask := image.NewNRGBA(img.Bounds())
			draw.Draw(mask, tc.region, image.White, image.Point{}, draw.Src)

			proc := &Processor{
				NewWidth:       tc.newWidth,
				NewHeight:      tc.newHeight,
				BlurRadius:     1,
				SobelThreshold: 4,
			}
			if tc.removal {
				proc.RMaskPath, proc.RMask = "mask.png", mask
			} else {
				proc.MaskPath, proc.Mask = "mask.png", mask
			}

			res := resizeThroughProcess(t, proc, img)

			wantWidth, wantHeight := tc.newWidth, tc.newHeight
			if wantWidth == 0 {
				wantWidth = 60
			}
			if wantHeight == 0 {
				wantHeight = 40
			}
			assert.Equal(t, image.Pt(wantWidth, wantHeight), res.Bounds().Size())
			assert.Equal(t, tc.wantSquarePixels, countColor(res, black))
		})
	}
}
