package caire

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

func Benchmark_Carver(b *testing.B) {
	sampleImg := filepath.Join("./testdata", "sample.jpg")
	f, err := os.Open(sampleImg)
	if err != nil {
		b.Fatalf("could not load sample image: %v", err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		b.Fatalf("error decoding image: %v", err)
	}
	img := p.imgToNRGBA(src)
	dx, dy := img.Bounds().Max.X, img.Bounds().Max.Y

	c := NewCarver(dx, dy)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		width, height := img.Bounds().Max.X, img.Bounds().Max.Y
		c = NewCarver(width, height)
		c.ComputeSeams(img, p)
		seams := c.FindLowestEnergySeams()
		img = c.RemoveSeam(img, seams, p.Debug)
	}

}
