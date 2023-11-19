package widget

import (
	"image"
	"math"
	"strings"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/text"
	"gioui.org/unit"
)

// stringSource is an immutable textSource with a fixed string
// value.
type stringSource struct {
	reader *strings.Reader
}

var _ textSource = stringSource{}

func newStringSource(str string) stringSource {
	return stringSource{
		reader: strings.NewReader(str),
	}
}

func (s stringSource) Changed() bool {
	return false
}

func (s stringSource) Size() int64 {
	return s.reader.Size()
}

func (s stringSource) ReadAt(b []byte, offset int64) (int, error) {
	return s.reader.ReadAt(b, offset)
}

// ReplaceRunes is unimplemented, as a stringSource is immutable.
func (s stringSource) ReplaceRunes(byteOffset, runeCount int64, str string) {
}

// Selectable displays selectable text.
type Selectable struct {
	// Alignment controls the alignment of the text.
	Alignment text.Alignment
	// MaxLines is the maximum number of lines of text to be displayed.
	MaxLines int
	// Truncator is the symbol to use at the end of the final line of text
	// if text was cut off. Defaults to "…" if left empty.
	Truncator string
	// WrapPolicy configures how displayed text will be broken into lines.
	WrapPolicy text.WrapPolicy
	// LineHeight controls the distance between the baselines of lines of text.
	// If zero, a sensible default will be used.
	LineHeight unit.Sp
	// LineHeightScale applies a scaling factor to the LineHeight. If zero, a
	// sensible default will be used.
	LineHeightScale float32
	initialized     bool
	source          stringSource
	// scratch is a buffer reused to efficiently read text out of the
	// textView.
	scratch      []byte
	lastValue    string
	text         textView
	focused      bool
	requestFocus bool
	dragging     bool
	dragger      gesture.Drag
	scroller     gesture.Scroll
	scrollOff    image.Point

	clicker gesture.Click
	// events is the list of events not yet processed.
	events []EditorEvent
	// prevEvents is the number of events from the previous frame.
	prevEvents int
}

// initialize must be called at the beginning of any exported method that
// manipulates text state. It ensures that the underlying text is safe to
// access.
func (l *Selectable) initialize() {
	if !l.initialized {
		l.source = newStringSource("")
		l.text.SetSource(l.source)
		l.initialized = true
	}
}

// Focus requests the input focus for the label.
func (l *Selectable) Focus() {
	l.requestFocus = true
}

// Focused returns whether the label is focused or not.
func (l *Selectable) Focused() bool {
	return l.focused
}

// paintSelection paints the contrasting background for selected text.
func (l *Selectable) paintSelection(gtx layout.Context, material op.CallOp) {
	l.initialize()
	if !l.focused {
		return
	}
	l.text.PaintSelection(gtx, material)
}

// paintText paints the text glyphs with the provided material.
func (l *Selectable) paintText(gtx layout.Context, material op.CallOp) {
	l.initialize()
	l.text.PaintText(gtx, material)
}

// SelectionLen returns the length of the selection, in runes; it is
// equivalent to utf8.RuneCountInString(e.SelectedText()).
func (l *Selectable) SelectionLen() int {
	l.initialize()
	return l.text.SelectionLen()
}

// Selection returns the start and end of the selection, as rune offsets.
// start can be > end.
func (l *Selectable) Selection() (start, end int) {
	l.initialize()
	return l.text.Selection()
}

// SetCaret moves the caret to start, and sets the selection end to end. start
// and end are in runes, and represent offsets into the editor text.
func (l *Selectable) SetCaret(start, end int) {
	l.initialize()
	l.text.SetCaret(start, end)
}

// SelectedText returns the currently selected text (if any) from the editor.
func (l *Selectable) SelectedText() string {
	l.initialize()
	l.scratch = l.text.SelectedText(l.scratch)
	return string(l.scratch)
}

// ClearSelection clears the selection, by setting the selection end equal to
// the selection start.
func (l *Selectable) ClearSelection() {
	l.initialize()
	l.text.ClearSelection()
}

// Text returns the contents of the label.
func (l *Selectable) Text() string {
	l.initialize()
	l.scratch = l.text.Text(l.scratch)
	return string(l.scratch)
}

// SetText updates the text to s if it does not already contain s. Updating the
// text will clear the selection unless the selectable already contains s.
func (l *Selectable) SetText(s string) {
	l.initialize()
	if l.lastValue != s {
		l.source = newStringSource(s)
		l.lastValue = s
		l.text.SetSource(l.source)
	}
}

// Truncated returns whether the text has been truncated by the text shaper to
// fit within available constraints.
func (l *Selectable) Truncated() bool {
	return l.text.Truncated()
}

// Layout clips to the dimensions of the selectable, updates the shaped text, configures input handling, and paints
// the text and selection rectangles. The provided textMaterial and selectionMaterial ops are used to set the
// paint material for the text and selection rectangles, respectively.
func (l *Selectable) Layout(gtx layout.Context, lt *text.Shaper, font font.Font, size unit.Sp, textMaterial, selectionMaterial op.CallOp) layout.Dimensions {
	l.initialize()
	l.text.LineHeight = l.LineHeight
	l.text.LineHeightScale = l.LineHeightScale
	l.text.Alignment = l.Alignment
	l.text.MaxLines = l.MaxLines
	l.text.Truncator = l.Truncator
	l.text.WrapPolicy = l.WrapPolicy
	l.text.Update(gtx, lt, font, size, l.handleEvents)
	dims := l.text.Dimensions()
	defer clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
	pointer.CursorText.Add(gtx.Ops)
	var keys key.Set
	if l.focused {
		const keyFilterAllArrows = "(ShortAlt)-(Shift)-[←,→,↑,↓]|(Shift)-[⏎,⌤]|(ShortAlt)-(Shift)-[⌫,⌦]|(Shift)-[⇞,⇟,⇱,⇲]|Short-[C,V,X,A]|Short-(Shift)-Z"
		keys = keyFilterAllArrows
	}
	key.InputOp{Tag: l, Keys: keys}.Add(gtx.Ops)
	if l.requestFocus {
		key.FocusOp{Tag: l}.Add(gtx.Ops)
		key.SoftKeyboardOp{Show: true}.Add(gtx.Ops)
	}
	l.requestFocus = false

	l.clicker.Add(gtx.Ops)
	l.dragger.Add(gtx.Ops)

	l.paintSelection(gtx, selectionMaterial)
	l.paintText(gtx, textMaterial)
	return dims
}

func (l *Selectable) handleEvents(gtx layout.Context) {
	// Flush events from before the previous Layout.
	n := copy(l.events, l.events[l.prevEvents:])
	l.events = l.events[:n]
	l.prevEvents = n
	oldStart, oldLen := min(l.text.Selection()), l.text.SelectionLen()
	l.processPointer(gtx)
	l.processKey(gtx)
	// Queue a SelectEvent if the selection changed, including if it went away.
	if newStart, newLen := min(l.text.Selection()), l.text.SelectionLen(); oldStart != newStart || oldLen != newLen {
		l.events = append(l.events, SelectEvent{})
	}
}

func (e *Selectable) processPointer(gtx layout.Context) {
	for _, evt := range e.clickDragEvents(gtx) {
		switch evt := evt.(type) {
		case gesture.ClickEvent:
			switch {
			case evt.Type == gesture.TypePress && evt.Source == pointer.Mouse,
				evt.Type == gesture.TypeClick && evt.Source != pointer.Mouse:
				prevCaretPos, _ := e.text.Selection()
				e.text.MoveCoord(image.Point{
					X: int(math.Round(float64(evt.Position.X))),
					Y: int(math.Round(float64(evt.Position.Y))),
				})
				e.requestFocus = true
				if evt.Modifiers == key.ModShift {
					start, end := e.text.Selection()
					// If they clicked closer to the end, then change the end to
					// where the caret used to be (effectively swapping start & end).
					if abs(end-start) < abs(start-prevCaretPos) {
						e.text.SetCaret(start, prevCaretPos)
					}
				} else {
					e.text.ClearSelection()
				}
				e.dragging = true

				// Process multi-clicks.
				switch {
				case evt.NumClicks == 2:
					e.text.MoveWord(-1, selectionClear)
					e.text.MoveWord(1, selectionExtend)
					e.dragging = false
				case evt.NumClicks >= 3:
					e.text.MoveStart(selectionClear)
					e.text.MoveEnd(selectionExtend)
					e.dragging = false
				}
			}
		case pointer.Event:
			release := false
			switch {
			case evt.Type == pointer.Release && evt.Source == pointer.Mouse:
				release = true
				fallthrough
			case evt.Type == pointer.Drag && evt.Source == pointer.Mouse:
				if e.dragging {
					e.text.MoveCoord(image.Point{
						X: int(math.Round(float64(evt.Position.X))),
						Y: int(math.Round(float64(evt.Position.Y))),
					})

					if release {
						e.dragging = false
					}
				}
			}
		}
	}
}

func (e *Selectable) clickDragEvents(gtx layout.Context) []event.Event {
	var combinedEvents []event.Event
	for _, evt := range e.clicker.Events(gtx) {
		combinedEvents = append(combinedEvents, evt)
	}
	for _, evt := range e.dragger.Events(gtx.Metric, gtx, gesture.Both) {
		combinedEvents = append(combinedEvents, evt)
	}
	return combinedEvents
}

func (e *Selectable) processKey(gtx layout.Context) {
	for _, ke := range gtx.Events(e) {
		switch ke := ke.(type) {
		case key.FocusEvent:
			e.focused = ke.Focus
		case key.Event:
			if !e.focused || ke.State != key.Press {
				break
			}
			e.command(gtx, ke)
		}
	}
}

func (e *Selectable) command(gtx layout.Context, k key.Event) {
	direction := 1
	if gtx.Locale.Direction.Progression() == system.TowardOrigin {
		direction = -1
	}
	moveByWord := k.Modifiers.Contain(key.ModShortcutAlt)
	selAct := selectionClear
	if k.Modifiers.Contain(key.ModShift) {
		selAct = selectionExtend
	}
	if k.Modifiers == key.ModShortcut {
		switch k.Name {
		// Copy or Cut selection -- ignored if nothing selected.
		case "C", "X":
			e.scratch = e.text.SelectedText(e.scratch)
			if text := string(e.scratch); text != "" {
				clipboard.WriteOp{Text: text}.Add(gtx.Ops)
			}
		// Select all
		case "A":
			e.text.SetCaret(0, e.text.Len())
		}
		return
	}
	switch k.Name {
	case key.NameUpArrow:
		e.text.MoveLines(-1, selAct)
	case key.NameDownArrow:
		e.text.MoveLines(+1, selAct)
	case key.NameLeftArrow:
		if moveByWord {
			e.text.MoveWord(-1*direction, selAct)
		} else {
			if selAct == selectionClear {
				e.text.ClearSelection()
			}
			e.text.MoveCaret(-1*direction, -1*direction*int(selAct))
		}
	case key.NameRightArrow:
		if moveByWord {
			e.text.MoveWord(1*direction, selAct)
		} else {
			if selAct == selectionClear {
				e.text.ClearSelection()
			}
			e.text.MoveCaret(1*direction, int(selAct)*direction)
		}
	case key.NamePageUp:
		e.text.MovePages(-1, selAct)
	case key.NamePageDown:
		e.text.MovePages(+1, selAct)
	case key.NameHome:
		e.text.MoveStart(selAct)
	case key.NameEnd:
		e.text.MoveEnd(selAct)
	}
}

// Events returns available text events.
func (l *Selectable) Events() []EditorEvent {
	events := l.events
	l.events = nil
	l.prevEvents = 0
	return events
}

// Regions returns visible regions covering the rune range [start,end).
func (l *Selectable) Regions(start, end int, regions []Region) []Region {
	l.initialize()
	return l.text.Regions(start, end, regions)
}
