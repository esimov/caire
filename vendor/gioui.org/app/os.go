// SPDX-License-Identifier: Unlicense OR MIT

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
	// Decorated reports whether window decorations are provided automatically.
	Decorated bool
	// decoHeight is the height of the fallback decoration for platforms such
	// as Wayland that may need fallback client-side decorations.
	decoHeight unit.Dp
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
// Note that mode can be changed programatically as well as by the user
// clicking on the minimize/maximize buttons on the window's title bar.
type WindowMode uint8

const (
	// Windowed is the normal window mode with OS specific window decorations.
	Windowed WindowMode = iota
	// Fullscreen is the full screen window mode.
	Fullscreen
	// Minimized is for systems where the window can be minimized to an icon.
	Minimized
	// Maximized is for systems where the window can be made to fill the available monitor area.
	Maximized
)

// Option changes the mode of a Window.
func (m WindowMode) Option() Option {
	return func(_ unit.Metric, cnf *Config) {
		cnf.Mode = m
	}
}

// String returns the mode name.
func (m WindowMode) String() string {
	switch m {
	case Windowed:
		return "windowed"
	case Fullscreen:
		return "fullscreen"
	case Minimized:
		return "minimized"
	case Maximized:
		return "maximized"
	}
	return ""
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

func (o Orientation) String() string {
	switch o {
	case AnyOrientation:
		return "any"
	case LandscapeOrientation:
		return "landscape"
	case PortraitOrientation:
		return "portrait"
	}
	return ""
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
	SetCursor(cursor pointer.Cursor)
	// Wakeup wakes up the event loop and sends a WakeupEvent.
	Wakeup()
	// Perform actions on the window.
	Perform(system.Action)
	// EditorStateChanged notifies the driver that the editor state changed.
	EditorStateChanged(old, new editorState)
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

func walkActions(actions system.Action, do func(system.Action)) {
	for a := system.Action(1); actions != 0; a <<= 1 {
		if actions&a != 0 {
			actions &^= a
			do(a)
		}
	}
}
