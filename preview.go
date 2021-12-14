package caire

// showPreview spawns a new Gio GUI window and updates its content with the resized image received from a channel.
func (p *Processor) showPreview(
	imgWorker <-chan worker,
	errChan chan<- error,
	guiParams struct {
		width  int
		height int
	},
) {
	var gui = newGui(guiParams.width, guiParams.height)
	gui.cp = p
	gui.proc.wrk = imgWorker

	for err := range gui.Run() {
		errChan <- err
	}
}
