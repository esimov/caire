package utils

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestUtils_ShouldDownloadImage(t *testing.T) {
	f, err := DownloadImage("https://raw.githubusercontent.com/esimov/caire/master/testdata/sample.jpg")
	if err != nil {
		t.Fatalf("could't download test file: %v", err)
	}

	if !strings.Contains(f.Name(), "tmp") {
		t.Errorf("The downloaded image should have been saved in a temporary folder")
	}
}

func TestUtils_ShouldBeValidUrl(t *testing.T) {
	ok := IsValidUrl("https://github.com/esimov/caire/")
	if !ok {
		t.Errorf("A valid URL should have been provided")
	}
}

func TestUtils_ShouldDetectValidFileType(t *testing.T) {
	sampleImg := filepath.Join("../testdata", "sample.jpg")

	ftype, err := DetectFileContentType(sampleImg)
	if err != nil {
		t.Fatalf("could not detect content type: %v", err)
	}

	if !strings.Contains(ftype.(string), "image") {
		t.Errorf("Content type expected to be of type image, got: %v", ftype)
	}
}
