package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"gioui.org/app"
	"github.com/esimov/caire"
	"github.com/esimov/caire/common"
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
	source         = flag.String("in", pipeName, "Source image")
	destination    = flag.String("out", pipeName, "Destination image")
	blurRadius     = flag.Int("blur", 4, "Blur radius")
	sobelThreshold = flag.Int("sobel", 2, "Sobel filter threshold")
	gaussianSigma  = flag.Float64("gaussian", 1.5, "Gaussian sigma for text detection")
	lowThreshold   = flag.Float64("cannyLow", 50, "Canny low threshold")
	highThreshold  = flag.Float64("cannyHigh", 300, "Canny high threshold")
	newWidth       = flag.Int("width", 0, "Desired image width")
	newHeight      = flag.Int("height", 0, "Desired image height")
	percentage     = flag.Bool("perc", false, "Reduce image by percentage")
	square         = flag.Bool("square", false, "Reduce image to square dimensions")
	debug          = flag.Bool("debug", false, "Show the seams")
	shapeType      = flag.String("shape", "circle", "Shape type used for debugging: circle|line")
	seamColor      = flag.String("color", "#ff0000", "Seam color")
	preview        = flag.Bool("preview", true, "Show GUI window")
	maskPath       = flag.String("mask", "", "File path to the mask used to retain specific image areas. \n"+
		"These areas won't be altered")
	rMaskPath       = flag.String("rmask", "", "File path to the mask used to force removing specific image areas.")
	faceDetect      = flag.Bool("faceDetect", false, "Activate face detection")
	textDetect      = flag.Bool("textDetect", false, "Activate text detection")
	textDetInterval = flag.Int("textDetInterval", 1, "The time interval after the text detector should be re-activated. \n"+
		"For a better performance this value can be increased, but to the detriment of accuracy")
	faceAngle = flag.Float64("angle", 0.0, "Face rotation angle")
	workers   = flag.Int("conc", runtime.NumCPU(), "Number of files to process concurrently")
)

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s", fmt.Sprintf(HelpBanner, Version))
		flag.PrintDefaults()
	}
	flag.Parse()

	proc := &caire.Processor{
		BlurRadius:         *blurRadius,
		SobelThreshold:     *sobelThreshold,
		GaussianSigma:      *gaussianSigma,
		CannyLowThreshold:  *lowThreshold,
		CannyHighThreshold: *highThreshold,
		NewWidth:           *newWidth,
		NewHeight:          *newHeight,
		Percentage:         *percentage,
		Square:             *square,
		Debug:              *debug,
		Preview:            *preview,
		FaceDetect:         *faceDetect,
		TextDetect:         *textDetect,
		TextDetInterval:    *textDetInterval,
		FaceAngle:          *faceAngle,
		SeamColor:          *seamColor,
		ShapeType:          caire.ShapeType(*shapeType),
	}

	if !(*newWidth > 0 || *newHeight > 0 || *percentage || *square) {
		flag.Usage()
		log.Fatalf("%s%s",
			common.DecorateText("\nPlease provide a width, height or percentage for image rescaling!", common.ErrorMessage),
			common.DefaultColor,
		)
	} else {
		op := &caire.Image{
			Src:       *source,
			Dst:       *destination,
			MaskPath:  *maskPath,
			RMaskPath: *rMaskPath,
			Workers:   *workers,
			PipeName:  pipeName,
		}

		if *preview {
			// Spawn the image resize process in a new goroutine to avoid blocking the Gio thread,
			// which should run in the main OS thread on operating systems like MacOS.
			go proc.Execute(op)
			app.Main()
		} else {
			proc.Execute(op)
		}
	}
}
