package editor

import . "github.com/itchyny/bed/core"

// UI defines the required user interface for the editor.
type UI interface {
	Init(chan<- Event, <-chan struct{}) error
	Run(map[Mode]*KeyManager)
	Height() int
	Redraw(State) error
	Close() error
}
