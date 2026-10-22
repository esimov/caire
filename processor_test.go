package caire

import (
	"image"
	"runtime"
	"sync"
	"testing"
	"time"

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
