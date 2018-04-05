package event

// Range of event
type Range struct {
	From Position
	To   Position
}

// Position ...
type Position interface {
	isPosition()
}

// Absolute is the absolute position of the buffer.
type Absolute struct {
	Offset int64
}

func (p Absolute) isPosition() {}

// Relative is the relative position of the buffer.
type Relative struct {
	Offset int64
}

func (p Relative) isPosition() {}

// End is the end of the buffer.
type End struct{}

func (p End) isPosition() {}

// VisualStart is the start position of visual selection.
type VisualStart struct{}

func (p VisualStart) isPosition() {}

// VisualEnd is the end position of visual selection.
type VisualEnd struct{}

func (p VisualEnd) isPosition() {}
