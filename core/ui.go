package core

// UI defines the required user interface for the editor.
type UI interface {
	Init(eventCh chan<- Event, quit <-chan struct{}) error
	Run(km map[Mode]*KeyManager)
	Height() int
	Redraw(state State) error
	Close() error
}
