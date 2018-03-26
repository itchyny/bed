package editor

import (
	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/key"
)

// UI defines the required user interface for the editor.
type UI interface {
	Init(chan<- Event) error
	Run(map[Mode]*key.Manager)
	Size() (int, int)
	Redraw(State) error
	Close() error
}
