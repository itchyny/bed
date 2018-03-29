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

// VisualStart is the start position of visual selection.
type VisualStart struct{}

func (v VisualStart) isPosition() {}

// VisualEnd is the end position of visual selection.
type VisualEnd struct{}

func (v VisualEnd) isPosition() {}
