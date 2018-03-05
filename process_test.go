package caire

import (
	"image"
	"testing"
)

func TestProcessor_Resize(t *testing.T) {
	reduceImageH(t)
	reduceImageV(t)
}

func reduceImageH(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newWidth := ImgWidth / 2
	p := &Processor{
		BlurRadius:     2,
		SobelThreshold: 10,
		NewWidth:       newWidth,
		NewHeight:      ImgHeight,
		Percentage:     false,
		Square:         false,
		Debug:          false,
		Scale:          false,
	}
	// Reduce image size horizontally
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

func reduceImageV(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	var c = NewCarver(img.Bounds().Dx(), img.Bounds().Dy())
	newHeight := ImgHeight / 2
	p := &Processor{
		BlurRadius:     2,
		SobelThreshold: 10,
		NewWidth:       ImgWidth,
		NewHeight:      newHeight,
		Percentage:     false,
		Square:         false,
		Debug:          false,
		Scale:          false,
	}
	// Reduce image size horizontally
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
