package core

// UI defines the required user interface for the editor.
type UI interface {
	Init(ch chan<- Event, quit <-chan struct{}) error
	Start(km map[Mode]*KeyManager) error
	Height() int
	Redraw(state State) error
	Close() error
}
