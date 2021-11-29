// SPDX-License-Identifier: Unlicense OR MIT

// package app implements platform specific windows
// and GPU contexts.
package app

import (
	"errors"
	"image"
	"image/color"

	"gioui.org/io/key"

	"gioui.org/gpu"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/unit"
)

type size struct {
	Width  unit.Value
	Height unit.Value
}

// errOutOfDate is reported when the GPU surface dimensions or properties no
// longer match the window.
var errOutOfDate = errors.New("app: GPU surface out of date")

// Config describes a Window configuration.
type Config struct {
	// Size is the window dimensions (Width, Height).
	Size image.Point
	// MaxSize is the window maximum allowed dimensions.
	MaxSize image.Point
	// MinSize is the window minimum allowed dimensions.
	MinSize image.Point
	// Title is the window title displayed in its decoration bar.
	Title string
	// WindowMode is the window mode.
	Mode WindowMode
	// StatusColor is the color of the Android status bar.
	StatusColor color.NRGBA
	// NavigationColor is the color of the navigation bar
	// on Android, or the address bar in browsers.
	NavigationColor color.NRGBA
	// Orientation is the current window orientation.
	Orientation Orientation
	// CustomRenderer is true when the window content is rendered by the
	// client.
	CustomRenderer bool
}

// ConfigEvent is sent whenever the configuration of a Window changes.
type ConfigEvent struct {
	Config Config
}

func (c *Config) apply(m unit.Metric, options []Option) {
	for _, o := range options {
		o(m, c)
	}
}

type wakeupEvent struct{}

// WindowMode is the window mode (WindowMode.Option sets it).
//
// Supported platforms are macOS, X11, Windows, Android and JS.
type WindowMode uint8

const (
	// Windowed is the normal window mode with OS specific window decorations.
	Windowed WindowMode = iota
	// Fullscreen is the full screen window mode.
	Fullscreen
)

func (m WindowMode) Option() Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Mode = m
	}
}

// Orientation is the orientation of the app (Orientation.Option sets it).
//
// Supported platforms are Android and JS.
type Orientation uint8

const (
	// AnyOrientation allows the window to be freely orientated.
	AnyOrientation Orientation = iota
	// LandscapeOrientation constrains the window to landscape orientations.
	LandscapeOrientation
	// PortraitOrientation constrains the window to portrait orientations.
	PortraitOrientation
)

func (o Orientation) Option() Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Orientation = o
	}
}

type frameEvent struct {
	system.FrameEvent

	Sync bool
}

type context interface {
	API() gpu.API
	RenderTarget() (gpu.RenderTarget, error)
	Present() error
	Refresh() error
	Release()
	Lock() error
	Unlock()
}

// Driver is the interface for the platform implementation
// of a window.
type driver interface {
	// SetAnimating sets the animation flag. When the window is animating,
	// FrameEvents are delivered as fast as the display can handle them.
	SetAnimating(anim bool)

	// ShowTextInput updates the virtual keyboard state.
	ShowTextInput(show bool)

	SetInputHint(mode key.InputHint)

	NewContext() (context, error)

	// ReadClipboard requests the clipboard content.
	ReadClipboard()
	// WriteClipboard requests a clipboard write.
	WriteClipboard(s string)

	// Configure the window.
	Configure([]Option)

	// SetCursor updates the current cursor to name.
	SetCursor(name pointer.CursorName)

	// Raise the window at the top.
	Raise()

	// Close the window.
	Close()
	// Wakeup wakes up the event loop and sends a WakeupEvent.
	Wakeup()

	// Maximize will make the window as large as possible, but keep the frame decorations.
	Maximize()
	// Center will place the window at monitor center.
	Center()
}

type windowRendezvous struct {
	in   chan windowAndConfig
	out  chan windowAndConfig
	errs chan error
}

type windowAndConfig struct {
	window  *callbacks
	options []Option
}

func newWindowRendezvous() *windowRendezvous {
	wr := &windowRendezvous{
		in:   make(chan windowAndConfig),
		out:  make(chan windowAndConfig),
		errs: make(chan error),
	}
	go func() {
		var main windowAndConfig
		var out chan windowAndConfig
		for {
			select {
			case w := <-wr.in:
				var err error
				if main.window != nil {
					err = errors.New("multiple windows are not supported")
				}
				wr.errs <- err
				main = w
				out = wr.out
			case out <- main:
			}
		}
	}()
	return wr
}

func (wakeupEvent) ImplementsEvent() {}
func (ConfigEvent) ImplementsEvent() {}
