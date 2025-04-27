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
	b.ResetTimer()

	img := imgToNRGBA(src)

	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	c := NewCarver(width, height)

	for i := 0; i < b.N; i++ {
		_, err := c.ComputeSeams(p, img)
		if err != nil {
			b.FailNow()
		}

		seams := c.FindLowestEnergySeams(p)
		img = c.RemoveSeam(img, seams, p.Debug)
	}

}
