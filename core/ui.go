package core

// UI defines the required user interface for the editor.
type UI interface {
	Init(ch chan<- Event) error
	Start() error
	Height() int
	Redraw(state State) error
	Close() error
}
