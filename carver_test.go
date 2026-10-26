package caire

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"testing"

	cimage "github.com/esimov/caire/image"
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

	// A local Processor keeps the test independent of the order the tests run in.
	proc := &Processor{NewWidth: imgWidth, NewHeight: imgHeight, BlurRadius: 1, SobelThreshold: 4}

	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.White}, image.Point{}, draw.Src)

	// Replace the pixel colors in a single row from 0xff to 0xdd, except for a gap
	// of three pixels around column 5. The line has a high energy and the gap has none,
	// so the lowest energy seam has to pass through the gap instead of column 0.
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	for x := 0; x < dx; x++ {
		if x >= 4 && x <= 6 {
			continue
		}
		img.Pix[(5*dx+x)*4+0] = 0xdd
		img.Pix[(5*dx+x)*4+1] = 0xdd
		img.Pix[(5*dx+x)*4+2] = 0xdd
		img.Pix[(5*dx+x)*4+3] = 0xdd
	}

	var c = NewCarver(dx, dy)
	for x := 0; x < imgWidth; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y

		c = NewCarver(width, height)
		_, err := c.ComputeSeams(proc, img)
		assert.NoError(t, err)

		les := c.FindLowestEnergySeams(proc)
		seams = append(seams, les)
	}

	for i := range seams {
		for s := range seams[i] {
			totalEnergySeams += seams[i][s].X
			if seams[i][s].Y == 5 {
				assert.InDelta(t, 5, seams[i][s].X, 1, "the seam should pass through the gap in the line")
			}
		}
	}
	assert.Greater(t, totalEnergySeams, 0)
}

func TestCarver_DetectVerticalEnergySeam(t *testing.T) {
	var seams [][]Seam
	var totalEnergySeams int

	// A local Processor keeps the test independent of the order the tests run in.
	proc := &Processor{NewWidth: imgWidth, NewHeight: imgHeight, BlurRadius: 1, SobelThreshold: 4}

	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{image.White}, image.Point{}, draw.Src)

	// Replace the pixel colors in a single column from 0xff to 0xdd, except for a gap
	// of three pixels around row 5. The line has a high energy and the gap has none,
	// so the lowest energy seam has to pass through the gap instead of column 0.
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	for y := 0; y < dy; y++ {
		if y >= 4 && y <= 6 {
			continue
		}
		img.Pix[5*4+(dx*y)*4+0] = 0xdd
		img.Pix[5*4+(dx*y)*4+1] = 0xdd
		img.Pix[5*4+(dx*y)*4+2] = 0xdd
		img.Pix[5*4+(dx*y)*4+3] = 0xff
	}

	var c = NewCarver(dx, dy)
	img = cimage.RotateImage90(img)
	for x := 0; x < imgHeight; x++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y

		c = NewCarver(width, height)
		_, err := c.ComputeSeams(proc, img)
		assert.NoError(t, err)

		les := c.FindLowestEnergySeams(proc)
		seams = append(seams, les)
	}

	// rotateImage90 turns column 5 into this row and the gap into columns 4 to 6.
	lineRow := dx - 1 - 5
	for i := range seams {
		for s := range seams[i] {
			totalEnergySeams += seams[i][s].X
			if seams[i][s].Y == lineRow {
				assert.InDelta(t, 5, seams[i][s].X, 1, "the seam should pass through the gap in the line")
			}
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
	img := cimage.ImgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	// Transform the image to a pixel array.
	pixels := cimage.RgbToGrayscale(img)

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
	img := cimage.ImgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	c := NewCarver(dx, dy)
	// Transform the image to a pixel array.
	pixels := cimage.RgbToGrayscale(img)

	sobel := cimage.SobelDetector(img, float64(p.SobelThreshold))

	err = cimage.Stackblur(img, sobel, uint32(p.BlurRadius))
	assert.NoError(t, err)

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
	img := cimage.ImgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	// Transform the image to a pixel array.
	pixels := cimage.RgbToGrayscale(img)
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

func TestCarver_CumulativeEnergyShouldMatchTheReference(t *testing.T) {
	// A pattern with energy in every column, including the two border columns.
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			v := uint8((x*37 + y*91 + x*y*13) % 256)
			img.SetNRGBA(x, y, color.NRGBA{R: v, G: v, B: v, A: 0xff})
		}
	}

	proc := &Processor{SobelThreshold: 0}
	c := NewCarver(imgWidth, imgHeight)
	energy, err := c.ComputeSeams(proc, img)
	assert.NoError(t, err)

	// Rebuild the table: M(x, y) = e(x, y) + min of the three cells above.
	e := func(x, y int) float64 {
		r, _, _, a := energy.At(x, y).RGBA()
		return float64(r) / float64(a)
	}
	ref := make([]float64, imgWidth*imgHeight)
	for x := 0; x < imgWidth; x++ {
		ref[x] = e(x, 0)
	}
	for y := 1; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			m := ref[x+(y-1)*imgWidth]
			if x > 0 {
				m = math.Min(m, ref[x-1+(y-1)*imgWidth])
			}
			if x < imgWidth-1 {
				m = math.Min(m, ref[x+1+(y-1)*imgWidth])
			}
			ref[x+y*imgWidth] = e(x, y) + m
		}
	}

	// The right column used to start from the left column's cumulative value.
	for i := range ref {
		assert.InDelta(t, ref[i], c.Points[i], 1e-9, "x=%d y=%d", i%imgWidth, i/imgWidth)
	}
}

func TestCarver_AddSeamOnTheLastColumn(t *testing.T) {
	// Row 0 has its seam on column 0 and row 1 has its seam on the last column.
	img := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	for x := 0; x < 3; x++ {
		img.SetNRGBA(x, 0, color.NRGBA{A: 0xff})
	}
	img.SetNRGBA(0, 1, color.NRGBA{R: 0x64, G: 0x64, B: 0x64, A: 0xff})
	img.SetNRGBA(1, 1, color.NRGBA{R: 0x64, G: 0x64, B: 0x64, A: 0xff})
	img.SetNRGBA(2, 1, color.NRGBA{R: 0xc8, G: 0xc8, B: 0xc8, A: 0xff})

	c := NewCarver(3, 2)
	seams := []Seam{{X: 0, Y: 0}, {X: 2, Y: 1}}
	res := c.AddSeam(img, seams, false)

	// The inserted pixel is the average of its left neighbor (0x64) and of
	// itself (0xc8), since there is no pixel to the right. It used to be
	// averaged with the right neighbor read on the previous row (0x00).
	got := color.NRGBAModel.Convert(res.At(2, 1)).(color.NRGBA)
	assert.Equal(t, color.NRGBA{R: 0x96, G: 0x96, B: 0x96, A: 0xff}, got)
}

func TestResize_EnlargeTransparentImageKeepsAlpha(t *testing.T) {
	tests := []struct {
		name                string
		newWidth, newHeight int
	}{
		{name: "width", newWidth: 7},
		{name: "height", newHeight: 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proc := &Processor{
				NewWidth:       tc.newWidth,
				NewHeight:      tc.newHeight,
				BlurRadius:     1,
				SobelThreshold: 4,
			}
			res, err := proc.Resize(image.NewNRGBA(image.Rect(0, 0, 5, 4)))
			assert.NoError(t, err)

			bounds := res.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, alpha := res.At(x, y).RGBA()
					assert.Zero(t, alpha, "alpha at (%d, %d)", x, y)
				}
			}
		})
	}
}

func TestCarver_AddSeamUsesPremultipliedAlpha(t *testing.T) {
	t.Run("transparent neighbors", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 4, 1))
		img.SetNRGBA(0, 0, color.NRGBA{G: 0xff})
		img.SetNRGBA(2, 0, color.NRGBA{B: 0xff})
		img.SetNRGBA(3, 0, color.NRGBA{R: 0xff, A: 0xff})

		res := NewCarver(4, 1).AddSeam(img, []Seam{{X: 1, Y: 0}}, false)
		got := color.NRGBAModel.Convert(res.At(1, 0)).(color.NRGBA)
		assert.Zero(t, got.A)
	})

	t.Run("transparent and opaque neighbors", func(t *testing.T) {
		img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
		img.SetNRGBA(0, 0, color.NRGBA{G: 0xff})
		img.SetNRGBA(2, 0, color.NRGBA{R: 0xc8, A: 0xff})

		res := NewCarver(3, 1).AddSeam(img, []Seam{{X: 1, Y: 0}}, false)
		got := color.NRGBAModel.Convert(res.At(1, 0)).(color.NRGBA)
		assert.InDelta(t, 0xc8, got.R, 1)
		assert.Zero(t, got.G)
		assert.Zero(t, got.B)
		assert.Equal(t, uint8(0x7f), got.A)
	})
}

func TestCarver_AddSeamKeepsOpaqueOutput(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 0xff})
	img.SetNRGBA(1, 0, color.NRGBA{R: 40, G: 50, B: 60, A: 0xff})
	img.SetNRGBA(2, 0, color.NRGBA{R: 90, G: 100, B: 110, A: 0xff})

	res := NewCarver(3, 1).AddSeam(img, []Seam{{X: 1, Y: 0}}, false)
	want := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	want.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 0xff})
	want.SetNRGBA(1, 0, color.NRGBA{R: 50, G: 60, B: 70, A: 0xff})
	want.SetNRGBA(2, 0, color.NRGBA{R: 40, G: 50, B: 60, A: 0xff})
	want.SetNRGBA(3, 0, color.NRGBA{R: 90, G: 100, B: 110, A: 0xff})

	assert.Equal(t, want.Pix, res.Pix)
}

func TestCarver_SobelShouldDetectEdgesInEveryColorChannel(t *testing.T) {
	// Left half green, right half blue. The edge between them is invisible
	// in the red channel, which used to be the only channel the detector read.
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	green := &image.Uniform{color.NRGBA{G: 0xff, A: 0xff}}
	blue := &image.Uniform{color.NRGBA{B: 0xff, A: 0xff}}
	draw.Draw(img, image.Rect(0, 0, imgWidth/2, imgHeight), green, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(imgWidth/2, 0, imgWidth, imgHeight), blue, image.Point{}, draw.Src)

	sobel := cimage.SobelDetector(img, 0)

	var energy int
	for i := 0; i < len(sobel.Pix); i += 4 {
		energy += int(sobel.Pix[i])
	}
	assert.NotZero(t, energy)
}
