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

func TestPathToFile_ShouldTruncateExistingOutput(t *testing.T) {
	dir := t.TempDir()
	inPath := filepath.Join(dir, "input.png")
	outPath := filepath.Join(dir, "output.png")
	freshPath := filepath.Join(dir, "fresh.png")
	writeTestImage(t, inPath, 16, 12)
	assert.NoError(t, os.WriteFile(outPath, make([]byte, 1<<20), 0644))

	op := &Image{PipeName: "-"}
	src, dst, err := op.pathToFile(inPath, outPath)
	assert.NoError(t, err)
	proc := &Processor{NewWidth: 12, BlurRadius: 1, SobelThreshold: 4}
	assert.NoError(t, proc.Process(src, dst))
	assert.NoError(t, src.(*os.File).Close())
	assert.NoError(t, dst.(*os.File).Close())

	freshSrc, err := os.Open(inPath)
	assert.NoError(t, err)
	freshDst, err := os.Create(freshPath)
	assert.NoError(t, err)
	freshProc := &Processor{NewWidth: 12, BlurRadius: 1, SobelThreshold: 4}
	assert.NoError(t, freshProc.Process(freshSrc, freshDst))
	assert.NoError(t, freshSrc.Close())
	assert.NoError(t, freshDst.Close())

	gotInfo, err := os.Stat(outPath)
	assert.NoError(t, err)
	wantInfo, err := os.Stat(freshPath)
	assert.NoError(t, err)
	assert.Equal(t, wantInfo.Size(), gotInfo.Size())
}
