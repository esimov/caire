package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/esimov/caire"
)

const HelpBanner = `
┌─┐┌─┐┬┬─┐┌─┐
│  ├─┤│├┬┘├┤
└─┘┴ ┴┴┴└─└─┘

Content aware image resize library.
    Version: %s

`

// Version indicates the current build version.
var Version string

var (
	// Flags
	source         = flag.String("in", "", "Source")
	destination    = flag.String("out", "", "Destination")
	blurRadius     = flag.Int("blur", 1, "Blur radius")
	sobelThreshold = flag.Int("sobel", 10, "Sobel filter threshold")
	newWidth       = flag.Int("width", 0, "New width")
	newHeight      = flag.Int("height", 0, "New height")
	percentage     = flag.Bool("perc", false, "Reduce image by percentage")
	square         = flag.Bool("square", false, "Reduce image to square dimensions")
	debug          = flag.Bool("debug", false, "Use debugger")
	scale          = flag.Bool("scale", false, "Proportional scaling")
	faceDetect     = flag.Bool("face", false, "Use face detection")
	faceAngle      = flag.Float64("angle", 0.0, "Plane rotated faces angle")
	cascade        = flag.String("cc", "", "Cascade classifier")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(HelpBanner, Version))
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(*source) == 0 || len(*destination) == 0 {
		log.Fatal("Usage: caire -in input.jpg -out out.jpg")
	}

	if *newWidth > 0 || *newHeight > 0 || *percentage || *square {
		fs, err := os.Stat(*source)
		if err != nil {
			log.Fatalf("Unable to open source: %v", err)
		}

		toProcess := make(map[string]string)

		p := &caire.Processor{
			BlurRadius:     *blurRadius,
			SobelThreshold: *sobelThreshold,
			NewWidth:       *newWidth,
			NewHeight:      *newHeight,
			Percentage:     *percentage,
			Square:         *square,
			Debug:          *debug,
			Scale:          *scale,
			FaceDetect:     *faceDetect,
			FaceAngle:      *faceAngle,
			Classifier:     *cascade,
		}
		switch mode := fs.Mode(); {
		case mode.IsDir():
			// Supported image files.
			extensions := []string{".jpg", ".png", ".jpeg", ".bmp", ".gif"}

			// Read source directory.
			files, err := ioutil.ReadDir(*source)
			if err != nil {
				log.Fatalf("Unable to read dir: %v", err)
			}
			// Read destination file or directory.
			dst, err := os.Stat(*destination)
			if err != nil {
				log.Fatalf("Unable to get dir stats: %v", err)
			}

			// Check if the image destination is a directory or a file.
			if dst.Mode().IsRegular() {
				log.Fatal("Please specify a directory as destination!")
			}
			output, err := filepath.Abs(*destination)
			if err != nil {
				log.Fatalf("Unable to get absolute path: %v", err)
			}

			// Range over all the image files and save them into a slice.
			var images []string
			for _, f := range files {
				ext := filepath.Ext(f.Name())
				for _, iex := range extensions {
					if ext == iex {
						images = append(images, f.Name())
					}
				}
			}

			// Process images from directory.
			for _, img := range images {
				// Get the file base name.
				name := strings.TrimSuffix(img, filepath.Ext(img))
				dir := strings.TrimRight(*source, "/")
				out := output + "/" + name + ".jpg"
				in := dir + "/" + img

				toProcess[in] = out
			}

		case mode.IsRegular():
			toProcess[*source] = *destination
		}

		for in, out := range toProcess {
			inFile, err := os.Open(in)
			if err != nil {
				log.Fatalf("Unable to open source file: %v", err)
			}

			outFile, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				log.Fatalf("Unable to open output file: %v", err)
			}

			s := new(spinner)
			s.start("Processing...")

			start := time.Now()
			err = p.Process(inFile, outFile)
			s.stop()

			if err == nil {
				fmt.Printf("\nRescaled in: \x1b[92m%.2fs\n\x1b[0m", time.Since(start).Seconds())
				fmt.Printf("\x1b[39mSaved as: \x1b[92m%s \n\n\x1b[0m", path.Base(out))
			} else {
				fmt.Printf("\nError rescaling image %s. Reason: %s\n", inFile.Name(), err.Error())
			}

			inFile.Close()
			outFile.Close()
		}
	} else {
		log.Fatal("\x1b[31mPlease provide a width, height or percentage for image rescaling!\x1b[39m")
	}

	caire.RemoveTempImage(caire.TempImage)
}

type spinner struct {
	stopChan chan struct{}
}

// Start process
func (s *spinner) start(message string) {
	s.stopChan = make(chan struct{}, 1)

	go func() {
		for {
			for _, r := range `-\|/` {
				select {
				case <-s.stopChan:
					return
				default:
					fmt.Printf("\r%s%s %c%s", message, "\x1b[92m", r, "\x1b[39m")
					time.Sleep(time.Millisecond * 100)
				}
			}
		}
	}()
}

// End process
func (s *spinner) stop() {
	s.stopChan <- struct{}{}
}
