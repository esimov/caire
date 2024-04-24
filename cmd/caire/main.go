package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"gioui.org/app"
	"github.com/esimov/caire"
	"github.com/esimov/caire/utils"
)

const HelpBanner = `
┌─┐┌─┐┬┬─┐┌─┐
│  ├─┤│├┬┘├┤
└─┘┴ ┴┴┴└─└─┘

Content aware image resize library.
    Version: %s

`

// pipeName indicates that stdin/stdout is being used as file names.
const pipeName = "-"

// Version indicates the current build version.
var Version string

var (
	// Flags
	source         = flag.String("in", pipeName, "Source")
	destination    = flag.String("out", pipeName, "Destination")
	blurRadius     = flag.Int("blur", 4, "Blur radius")
	sobelThreshold = flag.Int("sobel", 2, "Sobel filter threshold")
	newWidth       = flag.Int("width", 0, "New width")
	newHeight      = flag.Int("height", 0, "New height")
	percentage     = flag.Bool("perc", false, "Reduce image by percentage")
	square         = flag.Bool("square", false, "Reduce image to square dimensions")
	debug          = flag.Bool("debug", false, "Show the seams")
	shapeType      = flag.String("shape", "circle", "Shape type used for debugging: circle|line")
	seamColor      = flag.String("color", "#ff0000", "Seam color")
	preview        = flag.Bool("preview", true, "Show GUI window")
	maskPath       = flag.String("mask", "", "Mask file path for retaining area")
	rMaskPath      = flag.String("rmask", "", "Mask file path for removing area")
	faceDetect     = flag.Bool("face", false, "Use face detection")
	faceAngle      = flag.Float64("angle", 0.0, "Face rotation angle")
	workers        = flag.Int("conc", runtime.NumCPU(), "Number of files to process concurrently")
)

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(HelpBanner, Version))
		flag.PrintDefaults()
	}
	flag.Parse()

	proc := &caire.Processor{
		BlurRadius:     *blurRadius,
		SobelThreshold: *sobelThreshold,
		NewWidth:       *newWidth,
		NewHeight:      *newHeight,
		Percentage:     *percentage,
		Square:         *square,
		Debug:          *debug,
		Preview:        *preview,
		FaceDetect:     *faceDetect,
		FaceAngle:      *faceAngle,
		MaskPath:       *maskPath,
		RMaskPath:      *rMaskPath,
		ShapeType:      *shapeType,
		SeamColor:      *seamColor,
	}

	if !(*newWidth > 0 || *newHeight > 0 || *percentage || *square) {
		flag.Usage()
		log.Fatal(fmt.Sprintf("%s%s",
			utils.DecorateText("\nPlease provide a width, height or percentage for image rescaling!", utils.ErrorMessage),
			utils.DefaultColor,
		))
	} else {
		op := &caire.Ops{
			Src:      *source,
			Dst:      *destination,
			Workers:  *workers,
			PipeName: pipeName,
		}

		if *preview {
			// When the preview mode is activated we have to execute the resizing process
			// in a separate goroutine in order to not block the Gio thread,
			// which have to run on the main OS thread of the operating systems like MacOS.
			go proc.Execute(op)
			app.Main()
		} else {
			proc.Execute(op)
		}
	}
}
