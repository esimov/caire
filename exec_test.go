package caire

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// writeTestImage writes a PNG of the given size with a color gradient to path.
func writeTestImage(t *testing.T, path string, width, height int) {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 255 / width), G: uint8(y * 255 / height), B: 0x80, A: 0xff})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestExecute_DirectoryModeShouldNotShareTheProcessor(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	sizes := map[string][2]int{"a.png": {100, 40}, "b.png": {80, 30}, "c.png": {60, 20}}
	for name, s := range sizes {
		writeTestImage(t, filepath.Join(src, name), s[0], s[1])
	}

	// The percentage option converts NewWidth to a pixel size during the resize.
	// Without a copy of the Processor per file, the concurrent workers would
	// overwrite each other's size and change the shared Processor.
	proc := &Processor{Percentage: true, NewWidth: 20, BlurRadius: 1, SobelThreshold: 4}
	proc.Execute(&Image{Src: src, Dst: dst, Workers: 3, PipeName: "-"})

	for name, s := range sizes {
		f, err := os.Open(filepath.Join(dst, name))
		if !assert.NoError(t, err, name) {
			continue
		}
		out, _, err := image.Decode(f)
		f.Close()
		assert.NoError(t, err, name)
		assert.Equal(t, image.Pt(s[0]*8/10, s[1]), out.Bounds().Size(), name)
	}
	assert.Equal(t, 20, proc.NewWidth)
}
