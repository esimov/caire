package caire

import (
	"os"
)

// showPreview spawns a new Gio GUI window and updates its content with the resized image received from a channel.
func (p *Processor) showPreview(
	imgWorker <-chan worker,
	errChan chan<- error,
	guiParams struct {
		width  int
		height int
	},
) {
	var gui = NewGUI(guiParams.width, guiParams.height)
	gui.cp = p
	gui.proc.wrk = imgWorker

	// Run the Gio GUI app in a separate goroutine
	go func() {
		if err := gui.Run(); err != nil {
			errChan <- err
		}
		// It's important to call os.Exit(0) in order to terminate
		// the execution of the GUI app when pressing ESC key.
		os.Exit(0)
	}()
}
