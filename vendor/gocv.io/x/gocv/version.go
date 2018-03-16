package gocv

/*
#include <stdlib.h>
#include "version.h"
*/
import "C"

// GoCVVersion of this package, for display purposes.
const GoCVVersion = "0.10.0"

// Version returns the current golang package version
func Version() string {
	return GoCVVersion
}

// OpenCVVersion returns the current OpenCV lib version
func OpenCVVersion() string {
	return C.GoString(C.openCVVersion())
}
