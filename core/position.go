package core

// Position represents a coordinate.
type Position struct {
	X, Y int
}

// Up moves the position up.
func (p *Position) Up() {
	p.X = p.X - 1
}

// Down moves the position down.
func (p *Position) Down() {
	p.X = p.X + 1
}

// Left moves the position left.
func (p *Position) Left() {
	p.Y = p.Y - 1
}

// Right moves the position right.
func (p *Position) Right() {
	p.Y = p.Y + 1
}
