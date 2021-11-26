package caire

import (
	"image"
	"testing"
)

func TestResize_ShrinkImageWidth(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newWidth := ImgWidth / 2

	p.NewWidth = newWidth
	p.NewHeight = ImgHeight

	for x := 0; x < newWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	imgWidth := img.Bounds().Max.X

	if imgWidth != newWidth {
		t.Errorf("Resulted image width expected to be %v. Got %v", newWidth, imgWidth)
	}
}

func TestResize_ShrinkImageHeight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := ImgHeight / 2

	p.NewWidth = ImgWidth
	p.NewHeight = newHeight

	img = c.RotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.RemoveSeam(img, seams, p.Debug)
	}
	img = c.RotateImage270(img)
	imgHeight := img.Bounds().Max.Y

	if imgHeight != newHeight {
		t.Errorf("Resulted image height expected to be %v. Got %v", newHeight, imgHeight)
	}
}

func TestResize_EnlargeImageWidth(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	origImgWidth := img.Bounds().Dx()
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newWidth := ImgWidth * 2

	p.NewWidth = newWidth
	p.NewHeight = ImgHeight

	for x := 0; x < newWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.AddSeam(img, seams, p.Debug)
	}
	imgWidth := img.Bounds().Max.X - origImgWidth

	if imgWidth != newWidth {
		t.Errorf("Resulted image width expected to be %v. Got %v", newWidth, imgWidth)
	}
}

func TestResize_EnlargeImageHeight(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	origImgHeigth := img.Bounds().Dy()
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := ImgHeight * 2

	p.NewWidth = ImgWidth
	p.NewHeight = newHeight

	img = c.RotateImage90(img)
	for x := 0; x < newHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.AddSeam(img, seams, p.Debug)
	}
	img = c.RotateImage270(img)
	imgHeight := img.Bounds().Max.Y - origImgHeigth

	if imgHeight != newHeight {
		t.Errorf("Resulted image height expected to be %v. Got %v", newHeight, imgHeight)
	}
}
