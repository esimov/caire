package caire

import (
	"image"
	"runtime"
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
