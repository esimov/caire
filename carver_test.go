package caire

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"testing"

	"github.com/esimov/caire/utils"
	pigo "github.com/esimov/pigo/core"
	"github.com/stretchr/testify/assert"
)

const (
	imgWidth  = 10
	imgHeight = 10
)

var p *Processor

func init() {
	p = &Processor{
		NewWidth:       imgWidth,
		NewHeight:      imgHeight,
		BlurRadius:     1,
		SobelThreshold: 4,
		Percentage:     false,
		Square:         false,
		Debug:          false,
	}
}

func TestCarver_EnergySeamShouldNotBeDetected(t *testing.T) {
	assert := assert.New(t)

	var seams [][]Seam
	var totalEnergySeams int

	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()

	var c = NewCarver(dx, dy)
	for range imgWidth {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)

		_, err := c.ComputeSeams(p, img)
		assert.NoError(err)

		les := c.FindLowestEnergySeams(p)
		seams = append(seams, les)
	}

	for i := range seams {
		for s := range seams[i] {
			totalEnergySeams += seams[i][s].X
		}
	}
	assert.Equal(0, totalEnergySeams)
}

func TestCarver_DetectHorizontalEnergySeam(t *testing.T) {
	var seams [][]Seam
	var totalEnergySeams int

	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.White}, image.Point{}, draw.Src)

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
	for x := 0; x < imgWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y

		c = NewCarver(width, height)
		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		les := c.FindLowestEnergySeams(p)
		seams = append(seams, les)
	}

	for i := range seams {
		for s := range seams[i] {
			totalEnergySeams += seams[i][s].X
		}
	}
	assert.Greater(t, totalEnergySeams, 0)
}

func TestCarver_DetectVerticalEnergySeam(t *testing.T) {
	var seams [][]Seam
	var totalEnergySeams int

	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.White}, image.Point{}, draw.Src)

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
	img = rotateImage90(img)
	for x := 0; x < imgHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y

		c = NewCarver(width, height)
		_, err := c.ComputeSeams(p, img)
		assert.NoError(t, err)

		les := c.FindLowestEnergySeams(p)
		seams = append(seams, les)
	}

	for i := range seams {
		for s := range seams[i] {
			totalEnergySeams += seams[i][s].X
		}
	}
	assert.Greater(t, totalEnergySeams, 0)
}

func TestCarver_RemoveSeam(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	bounds := img.Bounds()

	// We choose to fill up the background with an uniform white color
	// and afterwards we replace the colors in a single row with lower intensity ones.
	draw.Draw(img, bounds, &image.Uniform{image.White}, image.Point{}, draw.Src)
	origImg := img

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	// Replace the pixels in row 5 with lower intensity colors.
	for x := 0; x < dx; x++ {
		img.Set(x, 5, color.RGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff})
	}

	c := NewCarver(dx, dy)
	_, err := c.ComputeSeams(p, img)
	assert.NoError(t, err)

	seams := c.FindLowestEnergySeams(p)
	img = c.RemoveSeam(img, seams, false)

	isEq := true
	// The test should pass if the detector correctly finds the row which pixel values are of lower intensity.
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
	assert.False(t, isEq)
}

func TestCarver_AddSeam(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	bounds := img.Bounds()

	// We choose to fill up the background with an uniform white color
	// Afterwards we'll replace the colors in a single row with lower intensity ones.
	draw.Draw(img, bounds, &image.Uniform{image.White}, image.Point{}, draw.Src)
	origImg := img

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	// Replace the pixels in row 5 with lower intensity colors.
	for x := 0; x < dx; x++ {
		img.Set(x, 5, color.RGBA{R: 0xdd, G: 0xdd, B: 0xdd, A: 0xff})
	}

	c := NewCarver(dx, dy)
	_, err := c.ComputeSeams(p, img)
	assert.NoError(t, err)

	seams := c.FindLowestEnergySeams(p)
	img = c.AddSeam(img, seams, false)

	dx, dy = img.Bounds().Dx(), img.Bounds().Dy()

	isEq := true
	// The test should pass if the detector correctly finds the row which has lower intensity colors.
	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			r0, g0, b0, _ := origImg.At(x, y).RGBA()
			r1, g1, b1, _ := img.At(x, y).RGBA()

			if r0>>8 != r1>>8 && g0>>8 != g1>>8 && b0>>8 != b1>>8 {
				isEq = false
			}
		}
	}
	assert.False(t, isEq)
}

func TestCarver_ComputeSeams(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	// We choose to fill up the background with an uniform white color
	// Afterwards we'll replace the colors in a single row with lower intensity ones.
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	// Replace the pixels in row 5 with lower intensity colors.
	for x := 0; x < dx; x++ {
		img.Pix[(5*dx+x)*4+0] = 0xdd
		img.Pix[(5*dx+x)*4+1] = 0xdd
		img.Pix[(5*dx+x)*4+2] = 0xdd
		img.Pix[(5*dx+x)*4+3] = 0xdd
	}

	c := NewCarver(dx, dy)
	_, err := c.ComputeSeams(p, img)
	assert.NoError(t, err)

	otherThenZero := findNonZeroValue(c.Points)

	assert.True(t, otherThenZero)
}

func TestCarver_ShouldDetectFace(t *testing.T) {
	p.FaceDetect = true

	sampleImg := filepath.Join("./testdata", "sample.jpg")
	f, err := os.Open(sampleImg)
	if err != nil {
		t.Fatalf("could not load sample image: %v", err)
	}
	defer f.Close()

	p.FaceDetector, err = p.FaceDetector.Unpack(cascadeFile)
	if err != nil {
		t.Fatalf("error unpacking the cascade file: %v", err)
	}

	src, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("error decoding image: %v", err)
	}
	img := imgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	// Transform the image to a pixel array.
	pixels := rgbToGrayscale(img)

	cParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     utils.Max(dx, dy),
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   dy,
			Cols:   dx,
			Dim:    dx,
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	faces := p.FaceDetector.RunCascade(cParams, p.FaceAngle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = p.FaceDetector.ClusterDetections(faces, 0.2)

	assert.Equal(t, 1, len(faces))
}

func TestCarver_ShouldNotRemoveFaceZone(t *testing.T) {
	p.FaceDetect = true
	p.BlurRadius = 10

	sampleImg := filepath.Join("./testdata", "sample.jpg")
	f, err := os.Open(sampleImg)
	if err != nil {
		t.Fatalf("could not load sample image: %v", err)
	}
	defer f.Close()

	p.FaceDetector, err = p.FaceDetector.Unpack(cascadeFile)
	if err != nil {
		t.Fatalf("error unpacking the cascade file: %v", err)
	}

	src, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("error decoding image: %v", err)
	}
	img := imgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	c := NewCarver(dx, dy)
	// Transform the image to a pixel array.
	pixels := rgbToGrayscale(img)

	sobel := c.SobelDetector(img, float64(p.SobelThreshold))
	img = c.StackBlur(sobel, uint32(p.BlurRadius))

	cParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     utils.Max(dx, dy),
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   dy,
			Cols:   dx,
			Dim:    dx,
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	faces := p.FaceDetector.RunCascade(cParams, p.FaceAngle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = p.FaceDetector.ClusterDetections(faces, 0.2)

	// Range over all the detected faces and draw a white rectangle mask over each of them.
	// We need to trick the sobel detector to consider them as important image parts.
	var rect image.Rectangle
	for _, face := range faces {
		if face.Q > 5.0 {
			rect = image.Rect(
				face.Col-face.Scale/2,
				face.Row-face.Scale/2,
				face.Col+face.Scale/2,
				face.Row+face.Scale/2,
			)
			draw.Draw(sobel, rect, &image.Uniform{image.White}, image.Point{}, draw.Src)
		}
	}
	_, err = c.ComputeSeams(p, img)
	assert.Error(t, err)

	seams := c.FindLowestEnergySeams(p)

	for _, seam := range seams {
		if seam.X >= rect.Min.X && seam.X <= rect.Max.X {
			t.Errorf("Carver shouldn't remove seams from face zone")
			break
		}
	}
}

func TestCarver_ShouldNotResizeWithFaceDistorsion(t *testing.T) {
	p.FaceDetect = true
	p.BlurRadius = 10
	p.NewHeight = 200

	sampleImg := filepath.Join("./testdata", "sample.jpg")
	f, err := os.Open(sampleImg)
	if err != nil {
		t.Fatalf("could not load sample image: %v", err)
	}
	defer f.Close()

	p.FaceDetector, err = p.FaceDetector.Unpack(cascadeFile)
	if err != nil {
		t.Fatalf("error unpacking the cascade file: %v", err)
	}

	src, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("error decoding image: %v", err)
	}
	img := imgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	// Transform the image to a pixel array.
	pixels := rgbToGrayscale(img)
	cParams := pigo.CascadeParams{
		MinSize:     100,
		MaxSize:     utils.Max(dx, dy),
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   dy,
			Cols:   dx,
			Dim:    dx,
		},
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	faces := p.FaceDetector.RunCascade(cParams, p.FaceAngle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces = p.FaceDetector.ClusterDetections(faces, 0.2)

	for _, face := range faces {
		if p.NewHeight < face.Scale {
			t.Errorf("Should not resize image without face deformation.")
		}
	}
}

// findNonZeroValue utility function to check if the slice contains values other then zeros.
func findNonZeroValue(points []float64) bool {
	var found = false
	for i := 0; i < len(points); i++ {
		if points[i] != 0 {
			found = true
		}
	}
	return found
}
