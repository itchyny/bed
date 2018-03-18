package editor

import . "github.com/itchyny/bed/common"

// UI defines the required user interface for the editor.
type UI interface {
	Init(chan<- Event) error
	Run(map[Mode]*KeyManager)
	Size() (int, int)
	Redraw(State) error
	Close() error
}
