package caire

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

var p *Processor

func init() {
	p = &Processor{
		NewWidth:       ImgWidth,
		NewHeight:      ImgHeight,
		BlurRadius:     1,
		SobelThreshold: 4,
		Percentage:     false,
		Square:         false,
		Debug:          false,
	}
}

func TestCarver_EnergySeamShouldNotBeDetected(t *testing.T) {
	var seams [][]Seam
	var totalEnergySeams int

	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()

	var c = NewCarver(dx, dy)
	for x := 0; x < ImgWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		les := c.FindLowestEnergySeams()
		seams = append(seams, les)
	}

	for i := 0; i < len(seams); i++ {
		for s := 0; s < len(seams[i]); s++ {
			totalEnergySeams += seams[i][s].X
		}
	}
	if totalEnergySeams != 0 {
		t.Errorf("Energy seam shouldn't been detected")
	}
}

func TestCarver_DetectHorizontalEnergySeam(t *testing.T) {
	var seams [][]Seam
	var totalEnergySeams int

	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.White}, image.ZP, draw.Src)

	// Replace the pixel colors in a single row from 0xff to 0xdd. 5 is an arbitrary value.
	// The seam detector should recognize that line as being of low energy density
	// and should perform the seam computation process.
	// This way we'll make sure, that the seam detector correctly detects one and only one line.
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	for x := 0; x < dx; x++ {
		img.Pix[(5*dx+x)*4+0] = 0xdd
		img.Pix[(5*dx+x)*4+1] = 0xdd
		img.Pix[(5*dx+x)*4+2] = 0xdd
		img.Pix[(5*dx+x)*4+3] = 0xdd
	}

	var c = NewCarver(dx, dy)
	for x := 0; x < ImgWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		les := c.FindLowestEnergySeams()
		seams = append(seams, les)
	}

	for i := 0; i < len(seams); i++ {
		for s := 0; s < len(seams[i]); s++ {
			totalEnergySeams += seams[i][s].X
		}
	}
	if totalEnergySeams == 0 {
		t.Errorf("The seam detector should have detected a horizontal energy seam")
	}
}

func TestCarver_DetectVerticalEnergySeam(t *testing.T) {
	var seams [][]Seam
	var totalEnergySeams int

	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.White}, image.ZP, draw.Src)

	// Replace the pixel colors in a single column from 0xff to 0xdd. 5 is an arbitrary value.
	// The seam detector should recognize that line as being of low energy density
	// and should perform the seam computation process.
	// This way we'll make sure, that the seam detector correctly detects one and only one line.
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	for y := 0; y < dy; y++ {
		img.Pix[5*4+(dx*y)*4+0] = 0xdd
		img.Pix[5*4+(dx*y)*4+1] = 0xdd
		img.Pix[5*4+(dx*y)*4+2] = 0xdd
		img.Pix[5*4+(dx*y)*4+3] = 0xff
	}

	var c = NewCarver(dx, dy)
	img = c.RotateImage90(img)
	for x := 0; x < ImgHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		les := c.FindLowestEnergySeams()
		seams = append(seams, les)
	}
	img = c.RotateImage270(img)

	for i := 0; i < len(seams); i++ {
		for s := 0; s < len(seams[i]); s++ {
			totalEnergySeams += seams[i][s].X
		}
	}
	if totalEnergySeams == 0 {
		t.Errorf("The seam detector should have detected a vertical energy seam")
	}
}

func TestCarver_RemoveSeam(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	bounds := img.Bounds()

	// We choose to fill up the background with an uniform white color
	// and afterwards we replace the colors in a single row with lower intensity ones.
	draw.Draw(img, bounds, &image.Uniform{image.White}, image.ZP, draw.Src)
	origImg := img

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	// Replace the pixels in row 5 with lower intensity colors.
	for x := 0; x < dx; x++ {
		img.Set(x, 5, color.RGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff})
	}

	c := NewCarver(dx, dy)
	c.ComputeSeams(img, p)
	seams := c.FindLowestEnergySeams()
	img = c.RemoveSeam(img, seams, false)

	isEq := true
	// The test should pass if the detector correctly finds the row wich pixel values are of lower intensity.
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			// In case the seam detector correctly recognize the modified line as of low importance
			// it should remove it, which means the new image width should be 1px less then the original image.
			r0, g0, b0, _ := origImg.At(x, y).RGBA()
			r1, g1, b1, _ := img.At(x, y).RGBA()

			if r0>>8 != r1>>8 && g0>>8 != g1>>8 && b0>>8 != b1>>8 {
				isEq = false
			}
		}
	}
	if isEq {
		t.Errorf("Seam should have been removed")
	}
}

func TestCarver_AddSeam(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, ImgWidth, ImgHeight))
	bounds := img.Bounds()

	// We choose to fill up the background with an uniform white color
	// Afterwards we'll replace the colors in a single row with lower intensity ones.
	draw.Draw(img, bounds, &image.Uniform{image.White}, image.ZP, draw.Src)
	origImg := img

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	// Replace the pixels in row 5 with lower intensity colors.
	for x := 0; x < dx; x++ {
		img.Set(x, 5, color.RGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff})
	}

	c := NewCarver(dx, dy)
	c.ComputeSeams(img, p)
	seams := c.FindLowestEnergySeams()
	img = c.AddSeam(img, seams, false)

	dx, dy = img.Bounds().Dx(), img.Bounds().Dy()

	isEq := true
	// The test should pass if the detector correctly finds the row wich has lower intensity colors.
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r0, g0, b0, _ := origImg.At(x, y).RGBA()
			r1, g1, b1, _ := img.At(x, y).RGBA()

			if r0>>8 != r1>>8 && g0>>8 != g1>>8 && b0>>8 != b1>>8 {
				isEq = false
			}
		}
	}
	if isEq {
		t.Errorf("Seam should have been added")
	}
}
