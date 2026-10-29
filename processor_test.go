package caire

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	cimage "github.com/esimov/caire/image"
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

	img = cimage.RotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		seams := c.FindLowestEnergySeams(p)
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	img = cimage.RotateImage270(img)
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

	img = cimage.RotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		seams := c.FindLowestEnergySeams(p)
		img = c.AddSeam(img, seams, p.Debug)
	}
	img = cimage.RotateImage270(img)
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
				proc.RMask = mask
			} else {
				proc.Mask = mask
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

func TestResize_NoPreviewShouldNotLeakGoroutines(t *testing.T) {
	proc := &Processor{
		NewWidth:       imgWidth,
		BlurRadius:     1,
		SobelThreshold: 4,
	}
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth*3, imgHeight))

	before := runtime.NumGoroutine()
	_, err := proc.Resize(img)
	assert.NoError(t, err)

	// Give the scheduler the chance to run any goroutine started by Resize.
	time.Sleep(50 * time.Millisecond)

	// Without a preview window nobody reads the GUI channel, so every seam
	// used to leave behind a goroutine which also kept its image in memory.
	assert.LessOrEqual(t, runtime.NumGoroutine(), before+1)
}

func TestResize_ReusedProcessorShouldNotKeepOldSeams(t *testing.T) {
	proc := &Processor{
		NewWidth:       imgWidth * 3,
		BlurRadius:     1,
		SobelThreshold: 4,
	}

	for i := 0; i < 2; i++ {
		img := image.NewNRGBA(image.Rect(0, 0, imgWidth*2, imgHeight))
		_, err := proc.Resize(img)
		assert.NoError(t, err)
	}
	// Enlarging by 10px records 10 seams. The seams of the first image
	// must not influence the second image.
	assert.Equal(t, imgWidth, len(proc.energySeams))
}

func TestResize_ShouldResizeBothAxesWhenCalledDirectly(t *testing.T) {
	proc := &Processor{
		NewWidth:       30,
		NewHeight:      20,
		BlurRadius:     1,
		SobelThreshold: 4,
	}
	img := image.NewNRGBA(image.Rect(0, 0, 40, 30))

	res, err := proc.Resize(img)
	assert.NoError(t, err)
	assert.Equal(t, image.Pt(30, 20), res.Bounds().Size())
}

func TestResize_ConcurrentProcessorsShouldBeIndependent(t *testing.T) {
	sizes := []struct{ width, height, newWidth int }{
		{30, 10, 20}, {40, 12, 25}, {50, 14, 35}, {60, 16, 45},
	}

	var wg sync.WaitGroup
	results := make([]image.Image, len(sizes))
	errs := make([]error, len(sizes))

	for i, s := range sizes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proc := &Processor{
				NewWidth:       s.newWidth,
				BlurRadius:     1,
				SobelThreshold: 4,
			}
			img := image.NewNRGBA(image.Rect(0, 0, s.width, s.height))
			results[i], errs[i] = proc.Resize(img)
		}()
	}
	wg.Wait()

	for i, s := range sizes {
		assert.NoError(t, errs[i])
		assert.Equal(t, image.Pt(s.newWidth, s.height), results[i].Bounds().Size())
	}
}

func TestResize_PercentageShouldNotChangeTheRequestedSize(t *testing.T) {
	proc := &Processor{
		Percentage:     true,
		NewWidth:       20,
		BlurRadius:     1,
		SobelThreshold: 4,
	}

	// Directory mode resizes every file with the same Processor.
	// The second image used to be shrunk to 20px instead of 20%.
	for i := 0; i < 2; i++ {
		img := image.NewNRGBA(image.Rect(0, 0, 100, 80))
		res, err := proc.Resize(img)
		assert.NoError(t, err)
		assert.Equal(t, image.Pt(80, 80), res.Bounds().Size())
	}
	assert.Equal(t, 20, proc.NewWidth)
}

func TestResize_ReusedProcessorShouldKeepTheMask(t *testing.T) {
	mask := image.NewNRGBA(image.Rect(0, 0, 60, 40))
	draw.Draw(mask, image.Rect(20, 10, 40, 30), image.White, image.Point{}, draw.Src)

	proc := &Processor{
		NewWidth:       50,
		BlurRadius:     1,
		SobelThreshold: 4,
		Mask:           mask,
	}

	// Each run carves the mask together with the image. The second run
	// used to read a mask of the first run's output size.
	for i := 0; i < 2; i++ {
		img := image.NewNRGBA(image.Rect(0, 0, 60, 40))
		res, err := proc.Resize(img)
		assert.NoError(t, err)
		assert.Equal(t, image.Pt(50, 40), res.Bounds().Size())
		assert.Same(t, mask, proc.Mask)
	}
}

func TestResize_ShouldRejectMismatchedMaskDimensions(t *testing.T) {
	tests := []struct {
		name    string
		proc    *Processor
		wantErr string
	}{
		{
			name: "protective mask",
			proc: &Processor{
				Mask: image.NewNRGBA(image.Rect(0, 0, 1, 2)),
			},
			wantErr: "mask dimensions 1x2 do not match input dimensions 4x2",
		},
		{
			name: "removal mask",
			proc: &Processor{
				RMask: image.NewNRGBA(image.Rect(0, 0, 1, 2)),
			},
			wantErr: "removal mask dimensions 1x2 do not match input dimensions 4x2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.proc.NewWidth = 2
			tc.proc.BlurRadius = 1
			tc.proc.SobelThreshold = 4

			_, err := tc.proc.Resize(image.NewNRGBA(image.Rect(0, 0, 4, 2)))
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("Resize() error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
