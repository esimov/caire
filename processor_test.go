package caire

import (
	"image"
	"testing"

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
