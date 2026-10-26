package caire

import (
	"image"
	"image/color"
	"image/draw"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// energyRow returns one character per pixel of a row: "#" for an edge, "." otherwise.
func energyRow(sobel *image.NRGBA, y int) string {
	var row strings.Builder
	for x := 0; x < sobel.Bounds().Dx(); x++ {
		if sobel.Pix[sobel.PixOffset(x, y)] > 0 {
			row.WriteByte('#')
		} else {
			row.WriteByte('.')
		}
	}
	return row.String()
}

func TestSobel_UniformImageShouldHaveNoEnergy(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Gray{Y: 0x80}}, image.Point{}, draw.Src)

	c := NewCarver(imgWidth, imgHeight)
	sobel := c.SobelDetector(img, 0)

	// The detector used to read past the end of the data in the last rows,
	// which marked the bottom of every image as an edge.
	for y := 0; y < imgHeight; y++ {
		assert.Equal(t, strings.Repeat(".", imgWidth), energyRow(sobel, y), "row %d", y)
	}
}

func TestSobel_EdgeShouldBeCenteredOnTheColorStep(t *testing.T) {
	// The left half is black and the right half is white.
	img := image.NewNRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(img, image.Rect(imgWidth/2, 0, imgWidth, imgHeight), image.White, image.Point{}, draw.Src)

	c := NewCarver(imgWidth, imgHeight)
	sobel := c.SobelDetector(img, 0)

	// Only the two columns on each side of the step are edges.
	// The window used to be anchored at the top left corner, which moved
	// every edge one pixel up and to the left.
	want := strings.Repeat(".", imgWidth/2-1) + "##" + strings.Repeat(".", imgWidth/2-1)
	for y := 0; y < imgHeight; y++ {
		assert.Equal(t, want, energyRow(sobel, y), "row %d", y)
	}
}
