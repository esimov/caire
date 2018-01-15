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

var (
	// Flags
	source         = flag.String("in", "", "Source")
	destination    = flag.String("out", "", "Destination")
	blurRadius     = flag.Int("blur", 2, "Blur radius")
	sobelThreshold = flag.Int("sobel", 50, "Sobel filter threshold")
	newWidth	= flag.Int("width", 0, "New width")
	newHeight	= flag.Int("height", 0, "New height")
	percentage	= flag.Int("percentage", 0, "Reduce by percentage")
)

func main() {
	flag.Parse()

	if len(*source) == 0 || len(*destination) == 0 {
		log.Fatal("Usage: caire -in input.jpg -out out.jpg")
	}

	if (*newWidth > 0 || *newHeight > 0 || *percentage > 0) {
		fs, err := os.Stat(*source)
		if err != nil {
			log.Fatalf("Unable to open source: %v", err)
		}

		toProcess := make(map[string]string)

		p := &caire.Carver{
			BlurRadius:     *blurRadius,
			SobelThreshold: *sobelThreshold,
			NewWidth: *newWidth,
			NewHeight: *newHeight,
			Percentage: *percentage,
		}

		switch mode := fs.Mode(); {
		case mode.IsDir():
			// Supported image files.
			extensions := []string{".jpg", ".png"}

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
				os.Exit(2)
			}
			output, err := filepath.Abs(filepath.Base(*destination))
			if err != nil {
				log.Fatalf("Unable to get absolute path: %v", err)
			}

			// Range over all the image files and save them into a slice.
			images := []string{}
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
				out := output + "/" + name + ".png"
				in := dir + "/" + img

				toProcess[in] = out
			}

		case mode.IsRegular():
			toProcess[*source] = *destination
		}

		for in, out := range toProcess {
			file, err := os.Open(in)
			if err != nil {
				log.Fatalf("Unable to open source file: %v", err)
			}
			defer file.Close()

			s := new(spinner)
			s.start("Processing...")

			start := time.Now()
			_, err = p.Process(file, out)
			s.stop()

			if err == nil {
				fmt.Printf("\nRescaled in: \x1b[92m%.2fs\n", time.Since(start).Seconds())
				fmt.Printf("\x1b[39mSaved as: \x1b[92m%s \n\n", path.Base(out))
			} else {
				fmt.Printf("\nError rescaling image: %s: %s", file.Name(), err.Error())
			}
		}
	} else {
		log.Fatal("\x1b[31mPlease provide a width, height or percentage for image rescaling!\x1b[39m")
	}
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
