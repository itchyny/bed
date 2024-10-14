package event

// Range of event
type Range struct {
	From Position
	To   Position
}

// Position ...
type Position interface{ add(int64) Position }

// Absolute is the absolute position of the buffer.
type Absolute struct{ Offset int64 }

func (p Absolute) add(offset int64) Position {
	return Absolute{p.Offset + offset}
}

// Relative is the relative position of the buffer.
type Relative struct{ Offset int64 }

func (p Relative) add(offset int64) Position {
	return Relative{p.Offset + offset}
}

// End is the end of the buffer.
type End struct{ Offset int64 }

func (p End) add(offset int64) Position {
	return End{p.Offset + offset}
}

// VisualStart is the start position of visual selection.
type VisualStart struct{ Offset int64 }

func (p VisualStart) add(offset int64) Position {
	return VisualStart{p.Offset + offset}
}

// VisualEnd is the end position of visual selection.
type VisualEnd struct{ Offset int64 }

func (p VisualEnd) add(offset int64) Position {
	return VisualEnd{p.Offset + offset}
}
