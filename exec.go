package caire

import (
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"slices"

	"github.com/esimov/caire/utils"
	"golang.org/x/term"
)

var (
	// imgFile holds the file being accessed, be it normal file or pipe name.
	imgFile *os.File

	// Common file related variable
	fs os.FileInfo
)

type Image struct {
	Src, Dst, PipeName string
	Workers            int
}

// result holds the relevant information about the resizing process and the generated image.
type result struct {
	path string
	err  error
}

func Resize(s SeamCarver, img *image.NRGBA) (image.Image, error) {
	return s.Resize(img)
}

// Execute executes the image resizing process.
// In case the preview mode is activated it will be invoked in a separate goroutine
// in order to avoid blocking the main OS thread. Otherwise it will be called normally.
func (p *Processor) Execute(img *Image) {
	var err error
	defaultMsg := fmt.Sprintf("%s %s",
		utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
		utils.DecorateText("⇢ resizing image (be patient, it may take a while)...", utils.DefaultMessage),
	)
	p.Spinner = utils.NewSpinner(defaultMsg, time.Millisecond*80)

	// Supported files
	validExtensions := []string{".jpg", ".png", ".jpeg", ".bmp", ".gif"}

	// Check if source path is a local image or URL.
	if utils.IsValidUrl(img.Src) {
		src, err := utils.DownloadImage(img.Src)
		if src != nil {
			defer os.Remove(src.Name())
		}

		if err != nil {
			log.Fatalf(
				utils.DecorateText("Failed to load the source image: %v", utils.ErrorMessage),
				utils.DecorateText(err.Error(), utils.DefaultMessage),
			)
		}
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

		imgFile = img
	} else {
		// Check if the source is a pipe name or a regular file.
		if img.Src == img.PipeName {
			fs, err = os.Stdin.Stat()
		} else {
			fs, err = os.Stat(img.Src)
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
		_, err := os.Stat(img.Dst)
		if err != nil {
			err = os.Mkdir(img.Dst, 0755)
			if err != nil {
				log.Fatalf(
					utils.DecorateText("Unable to get dir stats: %v\n", utils.ErrorMessage),
					utils.DecorateText(err.Error(), utils.DefaultMessage),
				)
			}
		}
		p.Preview = false

		// Limit the concurrently running workers to maxWorkers.
		if img.Workers <= 0 || img.Workers > runtime.NumCPU() {
			img.Workers = runtime.NumCPU()
		}

		// Process recursively the image files from the specified directory concurrently.
		ch := make(chan result)
		done := make(chan any)
		defer close(done)

		paths, errc := walkDir(done, img.Src, validExtensions)

		wg.Add(img.Workers)
		for range img.Workers {
			go func() {
				defer wg.Done()
				img.consumer(p, img.Dst, ch, done, paths)
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
			img.printOpStatus(res.path, err)
		}

		if err = <-errc; err != nil {
			fmt.Fprintf(os.Stderr, utils.DecorateText(err.Error(), utils.ErrorMessage))
		}

	case mode.IsRegular() || mode&os.ModeNamedPipe != 0: // check for regular files or pipe names
		ext := filepath.Ext(img.Dst)
		if !slices.Contains(validExtensions, ext) && img.Dst != img.PipeName {
			log.Fatalf(utils.DecorateText(fmt.Sprintf("%v file type not supported", ext), utils.ErrorMessage))
		}

		err = img.process(p, img.Src, img.Dst)
		img.printOpStatus(img.Dst, err)
	}
	if err == nil {
		fmt.Fprintf(os.Stderr, "\nExecution time: %s\n", utils.DecorateText(
			utils.FormatTime(time.Since(now)), utils.SuccessMessage),
		)
	}
}

// consumer reads the path names from the paths channel and calls the resizing processor against the source image.
func (img *Image) consumer(
	p *Processor,
	dest string,
	res chan<- result,
	done <-chan any,
	paths <-chan string,
) {
	for src := range paths {
		dst := filepath.Join(dest, filepath.Base(src))
		err := img.process(p, src, dst)

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

// processor calls the resizer method over the source image and returns the error in case exists.
func (img *Image) process(p *Processor, in, out string) error {
	var (
		successMsg string
		errorMsg   string
	)
	// Start the progress indicator.
	p.Spinner.Start()

	successMsg = fmt.Sprintf("%s %s %s",
		utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
		utils.DecorateText("⇢", utils.DefaultMessage),
		utils.DecorateText("the image has been resized successfully ✔", utils.SuccessMessage),
	)

	errorMsg = fmt.Sprintf("%s %s %s",
		utils.DecorateText("⚡ CAIRE", utils.StatusMessage),
		utils.DecorateText("resizing image failed...", utils.DefaultMessage),
		utils.DecorateText("✘", utils.ErrorMessage),
	)

	src, dst, err := img.pathToFile(in, out)
	if err != nil {
		p.Spinner.StopMsg = errorMsg
		return err
	}

	// Capture CTRL-C signal and restores back the cursor visibility.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		func() {
			p.Spinner.RestoreCursor()
			os.Remove(dst.(*os.File).Name())
			os.Exit(1)
		}()
	}()

	defer func() {
		if img, ok := src.(*os.File); ok {
			if err := img.Close(); err != nil {
				log.Printf("could not close the opened file: %v", err)
			}
		}
	}()

	defer func() {
		if img, ok := dst.(*os.File); ok {
			if err := img.Close(); err != nil {
				log.Printf("could not close the opened file: %v", err)
			}
		}
	}()

	if len(p.MaskPath) > 0 {
		mask, err := decodeImg(p.MaskPath)
		if err != nil {
			return fmt.Errorf("cannot decode image: %w", err)
		}
		p.Mask = dither(imgToNRGBA(mask))
		p.DebugMask = p.Mask
	}

	if len(p.RMaskPath) > 0 {
		rmask, err := decodeImg(p.RMaskPath)
		if err != nil {
			return fmt.Errorf("cannot decode image: %w", err)
		}
		p.RMask = dither(imgToNRGBA(rmask))
		p.DebugMask = p.RMask
	}

	err = p.Process(src, dst)
	if err != nil {
		// remove the generated image file in case of an error
		os.Remove(dst.(*os.File).Name())

		p.Spinner.StopMsg = errorMsg
		// Stop the progress indicator.
		p.Spinner.Stop()

		return err
	} else {
		p.Spinner.StopMsg = successMsg
		// Stop the progress indicator.
		p.Spinner.Stop()
	}

	return nil
}

// pathToFile converts the source and destination paths to readable and writable files.
func (img *Image) pathToFile(in, out string) (io.Reader, io.Writer, error) {
	var (
		src io.Reader
		dst io.Writer
		err error
	)
	// Check if the source path is a local image or URL.
	if utils.IsValidUrl(in) {
		src = imgFile
	} else {
		// Check if the source is a pipe name or a regular file.
		if in == img.PipeName {
			if term.IsTerminal(int(os.Stdin.Fd())) {
				return nil, nil, errors.New("`-` should be used with a pipe for stdin")
			}
			src = os.Stdin
		} else {
			src, err = os.Open(in)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to open the source file: %v", err)
			}
		}
	}

	// Check if the destination is a pipe name or a regular file.
	if out == img.PipeName {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			return nil, nil, errors.New("`-` should be used with a pipe for stdout")
		}
		dst = os.Stdout
	} else {
		dst, err = os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create the destination file: %v", err)
		}
	}

	return src, dst, nil
}

// printOpStatus displays the relevant information about the image resizing process.
func (img *Image) printOpStatus(fname string, err error) {
	if err != nil {
		log.Fatalf(
			utils.DecorateText("\nError resizing the image: %s", utils.ErrorMessage),
			utils.DecorateText(fmt.Sprintf("\n\tReason: %v\n", err.Error()), utils.DefaultMessage),
		)
	} else {
		if fname != img.PipeName {
			fmt.Fprintf(os.Stderr, "\nThe image has been saved as: %s %s\n\n",
				utils.DecorateText(filepath.Base(fname), utils.SuccessMessage),
				utils.DefaultColor,
			)
		}
	}
}

// walkDir starts a new goroutine to walk the specified directory tree
// in recursive manner and sends the path of each regular file to a new channel.
// It finishes in case the done channel is getting closed.
func walkDir(
	done <-chan any,
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
			if slices.Contains(srcExts, fx) {
				isFileSupported = true
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
