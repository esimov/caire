package utils

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestUtils_ShouldDownloadImage(t *testing.T) {
	f, err := DownloadImage("https://raw.githubusercontent.com/esimov/caire/master/testdata/sample.jpg")
	if err != nil {
		t.Fatalf("could't download test file: %v", err)
	}

	if !strings.Contains(f.Name(), "image") {
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

	ftype, err := DetectContentType(sampleImg)
	if err != nil {
		t.Fatalf("could not detect content type: %v", err)
	}

	if !strings.Contains(ftype.(string), "image") {
		t.Errorf("Content type expected to be of type image, got: %v", ftype)
	}
}

func TestDownloadImage_ShouldReturnTransportErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	f, err := DownloadImage(url)
	if f != nil {
		f.Close()
		t.Fatalf("DownloadImage() file = %q, want nil", f.Name())
	}
	if err == nil {
		t.Fatal("DownloadImage() error = nil, want a transport error")
	}
	if !strings.Contains(err.Error(), "unable to download image file from URI") {
		t.Fatalf("DownloadImage() error = %q, want URL download context", err)
	}
}
