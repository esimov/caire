package system

// Locale provides language information for the current system.
type Locale struct {
	// Language is the BCP-47 tag for the primary language of the system.
	Language string
	// Direction indicates the primary direction of text and layout
	// flow for the system.
	Direction TextDirection
}

const (
	axisShift = iota
	progressionShift
)

// TextDirection defines a direction for text flow.
type TextDirection byte

const (
	// LTR is left-to-right text.
	LTR TextDirection = TextDirection(Horizontal<<axisShift) | TextDirection(FromOrigin<<progressionShift)
	// RTL is right-to-left text.
	RTL TextDirection = TextDirection(Horizontal<<axisShift) | TextDirection(TowardOrigin<<progressionShift)
)

// Axis returns the axis of the text layout.
func (d TextDirection) Axis() TextAxis {
	return TextAxis((d & (1 << axisShift)) >> axisShift)
}

// Progression returns the way that the text flows relative to the origin.
func (d TextDirection) Progression() TextProgression {
	return TextProgression((d & (1 << progressionShift)) >> progressionShift)
}

func (d TextDirection) String() string {
	switch d {
	case RTL:
		return "RTL"
	default:
		return "LTR"
	}
}

// TextAxis defines the layout axis of text.
type TextAxis byte

const (
	// Horizontal indicates text that flows along the X axis.
	Horizontal TextAxis = iota
	// Vertical indicates text that flows along the Y axis.
	Vertical
)

// TextProgression indicates how text flows along an axis relative to the
// origin. For these purposes, the origin is defined as the upper-left
// corner of coordinate space.
type TextProgression byte

const (
	// FromOrigin indicates text that flows along its axis away from the
	// origin (upper left corner).
	FromOrigin TextProgression = iota
	// TowardOrigin indicates text that flows along its axis towards the
	// origin (upper left corner).
	TowardOrigin
)
