package gocv

import (
	"strings"
	"testing"
)

func TestVersions(t *testing.T) {
	ocvv := OpenCVVersion()

	if !strings.Contains(ocvv, "3.4") {
		t.Error("Wrong version of OpenCV:", ocvv)
	}

	v := Version()

	if v != GoCVVersion {
		t.Error("Wrong version of GoCV")
	}
}
