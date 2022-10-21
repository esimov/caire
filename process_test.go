package caire

import (
	"image"
	"testing"
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
		c.ComputeSeams(p, img)
		seams := c.FindLowestEnergySeams(p)
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	imgWidth := img.Bounds().Max.X

	if imgWidth != newWidth {
		t.Errorf("Resulted image width expected to be %v. Got %v", newWidth, imgWidth)
	}
}

func TestResize_ShrinkImageHeight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := imgHeight / 2

	p.NewWidth = imgWidth
	p.NewHeight = newHeight

	img = c.RotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(p, img)
		seams := c.FindLowestEnergySeams(p)
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	img = c.RotateImage270(img)
	imgHeight := img.Bounds().Max.Y

	if imgHeight != newHeight {
		t.Errorf("Resulted image height expected to be %v. Got %v", newHeight, imgHeight)
	}
}

func TestResize_EnlargeImageWidth(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	origImgWidth := img.Bounds().Dx()
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newWidth := imgWidth * 2

	p.NewWidth = newWidth
	p.NewHeight = imgHeight

	for x := 0; x < newWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(p, img)
		seams := c.FindLowestEnergySeams(p)
		img = c.AddSeam(img, seams, p.Debug)
	}
	imgWidth := img.Bounds().Max.X - origImgWidth

	if imgWidth != newWidth {
		t.Errorf("Resulted image width expected to be %v. Got %v", newWidth, imgWidth)
	}
}

func TestResize_EnlargeImageHeight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	origImgHeigth := img.Bounds().Dy()
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := imgHeight * 2

	p.NewWidth = imgWidth
	p.NewHeight = newHeight

	img = c.RotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(p, img)
		seams := c.FindLowestEnergySeams(p)
		img = c.AddSeam(img, seams, p.Debug)
	}
	img = c.RotateImage270(img)
	imgHeight := img.Bounds().Max.Y - origImgHeigth

	if imgHeight != newHeight {
		t.Errorf("Resulted image height expected to be %v. Got %v", newHeight, imgHeight)
	}
}
