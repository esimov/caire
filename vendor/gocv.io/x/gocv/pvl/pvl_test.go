package pvl

import (
	"testing"
)

func TestFaceDetector(t *testing.T) {
	fd := NewFaceDetector()
	fd.Close()
}

func TestFaceRecognizer(t *testing.T) {
	fr := NewFaceRecognizer()
	defer fr.Close()
	if !fr.Empty() {
		t.Error("Invalid Face Recognizer")
	}

	if fr.GetNumRegisteredPersons() != 0 {
		t.Error("Invalid Face Recognizer registered persons")
	}
}
