package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/esimov/caire"
	"golang.org/x/crypto/ssh/terminal"
)

const HelpBanner = `
┌─┐┌─┐┬┬─┐┌─┐
│  ├─┤│├┬┘├┤
└─┘┴ ┴┴┴└─└─┘

Content aware image resize library.
    Version: %s

`

// PipeName is the file name that indicates stdin/stdout is being used.
const PipeName = "-"

// Version indicates the current build version.
var Version string

// Supported image files.
var extensions = []string{".jpg", ".png", ".jpeg", ".bmp", ".gif"}

var (
	// Flags
	source         = flag.String("in", PipeName, "Source")
	destination    = flag.String("out", PipeName, "Destination")
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
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf(HelpBanner, Version))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *newWidth > 0 || *newHeight > 0 || *percentage || *square {
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
		process(p, *destination, *source)
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
					fmt.Fprintf(os.Stderr, "\r%s%s %c%s", message, "\x1b[92m", r, "\x1b[39m")
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

func process(p *caire.Processor, dstname, srcname string) {
	var src io.Reader
	if srcname == PipeName {
		if terminal.IsTerminal(int(os.Stdin.Fd())) {
			log.Fatalln("`-` should be used with a pipe for stdin")
		}
		src = os.Stdin
	} else {
		srcinfo, err := os.Stat(srcname)
		if err != nil {
			log.Fatalf("Unable to open source: %v", err)
		}

		if srcinfo.IsDir() {
			dstinfo, err := os.Stat(dstname)
			if err != nil {
				log.Fatalf("Unable to get dir stats: %v", err)
			}
			if !dstinfo.IsDir() {
				log.Fatalf("Please specify a directory as destination!")
			}

			files, err := ioutil.ReadDir(srcname)
			if err != nil {
				log.Fatalf("Unable to read dir: %v", err)
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

				process(p, filepath.Join(dstname, name+".jpg"), filepath.Join(srcname, img))
			}
			return
		}

		f, err := os.Open(srcname)
		if err != nil {
			log.Fatalf("Unable to open source file: %v", err)
		}
		defer f.Close()
		src = f
	}

	var dst io.Writer
	if dstname == PipeName {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			log.Fatalln("`-` should be used with a pipe for stdout")
		}
		dst = os.Stdout
	} else {
		f, err := os.OpenFile(dstname, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.Fatalf("Unable to open output file: %v", err)
		}
		defer f.Close()
		dst = f
	}

	s := new(spinner)
	s.start("Processing...")

	start := time.Now()
	err := p.Process(src, dst)
	s.stop()

	if err == nil {
		log.Printf("\nRescaled in: \x1b[92m%.2fs\n\x1b[0m", time.Since(start).Seconds())
		if dstname != PipeName {
			log.Printf("\x1b[39mSaved as: \x1b[92m%s \n\n\x1b[0m", path.Base(dstname))
		}
	} else {
		log.Printf("\nError rescaling image %s. Reason: %s\n", srcname, err.Error())
	}
}
