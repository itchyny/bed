package editor

import (
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/key"
	"github.com/itchyny/bed/mode"
	"github.com/itchyny/bed/state"
)

// UI defines the required user interface for the editor.
type UI interface {
	Init(chan<- event.Event) error
	Run(map[mode.Mode]*key.Manager)
	Size() (int, int)
	Redraw(state.State) error
	Close() error
}
