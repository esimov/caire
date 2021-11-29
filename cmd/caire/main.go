package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/esimov/caire"
	"github.com/esimov/caire/utils"
	"golang.org/x/term"
)

const HelpBanner = `
┌─┐┌─┐┬┬─┐┌─┐
│  ├─┤│├┬┘├┤
└─┘┴ ┴┴┴└─└─┘

Content aware image resize library.
    Version: %s

`

// pipeName is the file name that indicates stdin/stdout is being used.
const pipeName = "-"

// maxWorkers sets the maximum number of concurrently running workers.
const maxWorkers = 20

// result holds the relevant information about the resizing process and the generated image.
type result struct {
	path string
	err  error
}

var (
	// imgurl holds the file being accessed be it normal file or pipe name.
	imgurl *os.File
	// spinner used to instantiate and call the progress indicator.
	spinner *utils.Spinner
)

// Version indicates the current build version.
var Version string

var (
	// Flags
	source         = flag.String("in", pipeName, "Source")
	destination    = flag.String("out", pipeName, "Destination")
	blurRadius     = flag.Int("blur", 1, "Blur radius")
	sobelThreshold = flag.Int("sobel", 10, "Sobel filter threshold")
	newWidth       = flag.Int("width", 0, "New width")
	newHeight      = flag.Int("height", 0, "New height")
	percentage     = flag.Bool("perc", false, "Reduce image by percentage")
	square         = flag.Bool("square", false, "Reduce image to square dimensions")
	debug          = flag.Bool("debug", false, "Use debugger")
	preview        = flag.Bool("preview", true, "Show preview window")
	faceDetect     = flag.Bool("face", false, "Use face detection")
	faceAngle      = flag.Float64("angle", 0.0, "Face rotation angle")
	workers        = flag.Int("conc", runtime.NumCPU(), "Number of files to process concurrently")

	// Common file related variable
	fs os.FileInfo
)

func main() {
	var err error

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
	}

	defaultMsg := fmt.Sprintf("%s %s",
		utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
		utils.DecorateText("⇢ image resizing in progress (be patient, it may take a while)...", utils.DefaultMessage),
	)

	spinner = utils.NewSpinner(defaultMsg, time.Millisecond*80)

	if *newWidth > 0 || *newHeight > 0 || *percentage || *square {
		// Supported files
		validExtensions := []string{".jpg", ".png", ".jpeg", ".bmp", ".gif"}

		// Check if source path is a local image or URL.
		if utils.IsValidUrl(*source) {
			src, err := utils.DownloadImage(*source)
			defer src.Close()
			defer os.Remove(src.Name())

			fs, err = src.Stat()
			if err != nil {
				log.Fatalf(
					utils.DecorateText("Failed to load the source image: %v", utils.ErrorMessage),
					utils.DecorateText(err.Error(), utils.DefaultMessage),
				)
			}
			img, err := os.Open(src.Name())
			if err != nil {
				log.Fatalf(
					utils.DecorateText("Unable to open the temporary image file: %v", utils.ErrorMessage),
					utils.DecorateText(err.Error(), utils.DefaultMessage),
				)
			}
			imgurl = img
		} else {
			// Check if the source is a pipe name or a regular file.
			if *source == pipeName {
				fs, err = os.Stdin.Stat()
			} else {
				fs, err = os.Stat(*source)
			}
			if err != nil {
				log.Fatalf(
					utils.DecorateText("Failed to load the source image: %v", utils.ErrorMessage),
					utils.DecorateText(err.Error(), utils.DefaultMessage),
				)
			}
		}

		now := time.Now()

		switch mode := fs.Mode(); {
		case mode.IsDir():
			var wg sync.WaitGroup
			// Read destination file or directory.
			_, err := os.Stat(*destination)
			if err != nil {
				err = os.Mkdir(*destination, 0755)
				if err != nil {
					log.Fatalf(
						utils.DecorateText("Unable to get dir stats: %v\n", utils.ErrorMessage),
						utils.DecorateText(err.Error(), utils.DefaultMessage),
					)
				}
			}
			proc.Preview = false

			// Limit the concurrently running workers to maxWorkers.
			if *workers <= 0 || *workers > maxWorkers {
				*workers = runtime.NumCPU()
			}

			// Process recursively the image files from the specified directory concurrently.
			ch := make(chan result)
			done := make(chan interface{})
			defer close(done)

			paths, errc := walkDir(done, *source, validExtensions)

			wg.Add(*workers)
			for i := 0; i < *workers; i++ {
				go func() {
					defer wg.Done()
					consumer(done, paths, *destination, proc, ch)
				}()
			}

			// Close the channel after the values are consumed.
			go func() {
				defer close(ch)
				wg.Wait()
			}()

			// Consume the channel values.
			for res := range ch {
				if res.err != nil {
					err = res.err
				}
				printStatus(res.path, res.err)
			}

			if err = <-errc; err != nil {
				fmt.Fprintf(os.Stderr, utils.DecorateText(err.Error(), utils.ErrorMessage))
			}

		case mode.IsRegular() || mode&os.ModeNamedPipe != 0: // check for regular files or pipe names
			ext := filepath.Ext(*destination)
			if !isValidExtension(ext, validExtensions) && *destination != pipeName {
				log.Fatalf(utils.DecorateText(fmt.Sprintf("%v file type not supported", ext), utils.ErrorMessage))
			}

			err = processor(*source, *destination, proc)
			printStatus(*destination, err)
		}
		if err == nil {
			fmt.Fprintf(os.Stderr, "\nExecution time: %s\n", utils.DecorateText(fmt.Sprintf("%s", utils.FormatTime(time.Since(now))), utils.SuccessMessage))
		}
	} else {
		flag.Usage()
		log.Fatal(fmt.Sprintf("%s%s",
			utils.DecorateText("\nPlease provide a width, height or percentage for image rescaling!", utils.ErrorMessage),
			utils.DefaultColor,
		))
	}
}

// walkDir starts a goroutine to walk the specified directory tree in recursive manner
// and send the path of each regular file on the string channel.
// It sends the result of the walk on the error channel.
// It terminates in case done channel is closed.
func walkDir(
	done <-chan interface{},
	src string,
	srcExts []string,
) (<-chan string, <-chan error) {
	pathChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		// Close the paths channel after Walk returns.
		defer close(pathChan)

		errChan <- filepath.Walk(src, func(path string, f os.FileInfo, err error) error {
			isFileSupported := false
			if err != nil {
				return err
			}
			if !f.Mode().IsRegular() {
				return nil
			}

			// Get the file base name.
			fx := filepath.Ext(f.Name())
			for _, ext := range srcExts {
				if ext == fx {
					isFileSupported = true
					break
				}
			}

			if isFileSupported {
				select {
				case <-done:
					return errors.New("directory walk cancelled")
				case pathChan <- path:
				}
			}
			return nil
		})
	}()
	return pathChan, errChan
}

// consumer reads the path names from the paths channel and
// calls the resizing processor against the source image
// then sends the results on a new channel.
func consumer(
	done <-chan interface{},
	paths <-chan string,
	dest string,
	proc *caire.Processor,
	res chan<- result,
) {
	for src := range paths {
		dst := filepath.Join(dest, filepath.Base(src))
		err := processor(src, dst, proc)

		select {
		case <-done:
			return
		case res <- result{
			path: src,
			err:  err,
		}:
		}
	}
}

// processor calls the resizer method over the source image and
// returns the error in case exists, otherwise nil.
func processor(in, out string, proc *caire.Processor) error {
	var (
		successMsg string
		errorMsg   string
	)
	// Start the progress indicator.
	spinner.Start()

	successMsg = fmt.Sprintf("%s %s %s",
		utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
		utils.DecorateText("⇢", utils.DefaultMessage),
		utils.DecorateText("the image has been resized sucessfully ✔", utils.SuccessMessage),
	)

	errorMsg = fmt.Sprintf("%s %s %s",
		utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
		utils.DecorateText("resizing image failed...", utils.DefaultMessage),
		utils.DecorateText("✘", utils.ErrorMessage),
	)

	src, dst, err := pathToFile(in, out)
	if err != nil {
		spinner.StopMsg = errorMsg
		return err
	}

	// Capture CTRL-C signal and restore the cursor visibility back.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		func() {
			spinner.RestoreCursor()
			os.Remove(dst.(*os.File).Name())
			os.Exit(1)
		}()
	}()

	defer src.(*os.File).Close()
	defer dst.(*os.File).Close()

	err = proc.Process(src, dst)
	if err != nil {
		// remove the generated image file in case of an error
		os.Remove(dst.(*os.File).Name())

		spinner.StopMsg = errorMsg
		// Stop the progress indicator.
		spinner.Stop()

		return err
	} else {
		spinner.StopMsg = successMsg
		// Stop the progress indicator.
		spinner.Stop()
	}

	return nil
}

// pathToFile converts the source and destination paths to readable and writable files.
func pathToFile(in, out string) (io.Reader, io.Writer, error) {
	var (
		src io.Reader
		dst io.Writer
		err error
	)
	// Check if the source path is a local image or URL.
	if utils.IsValidUrl(in) {
		src = imgurl
	} else {
		// Check if the source is a pipe name or a regular file.
		if in == pipeName {
			if term.IsTerminal(int(os.Stdin.Fd())) {
				return nil, nil, errors.New("`-` should be used with a pipe for stdin")
			}
			src = os.Stdin
		} else {
			src, err = os.Open(in)
			if err != nil {
				return nil, nil, errors.New(
					fmt.Sprintf("unable to open the source file: %v", err),
				)
			}
		}
	}

	// Check if the destination is a pipe name or a regular file.
	if out == pipeName {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			return nil, nil, errors.New("`-` should be used with a pipe for stdout")
		}
		dst = os.Stdout
	} else {
		dst, err = os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			return nil, nil, errors.New(
				fmt.Sprintf("unable to create the destination file: %v", err),
			)
		}
	}
	return src, dst, nil
}

// printStatus displays the relavant information about the image resizing process.
func printStatus(fname string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr,
			utils.DecorateText("\nError resizing the image: %s", utils.ErrorMessage),
			utils.DecorateText(fmt.Sprintf("\n\tReason: %v\n", err.Error()), utils.DefaultMessage),
		)
	} else {
		if fname != pipeName {
			fmt.Fprintf(os.Stderr, fmt.Sprintf("\nThe resized image has been saved as: %s %s\n\n",
				utils.DecorateText(filepath.Base(fname), utils.SuccessMessage),
				utils.DefaultColor,
			))
		}
	}
}

// isValidExtension checks for the supported extensions.
func isValidExtension(ext string, extensions []string) bool {
	for _, ex := range extensions {
		if ex == ext {
			return true
		}
	}
	return false
}
