package editor

import (
	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/layout"
	"github.com/itchyny/bed/state"
)

// Manager defines the required window manager interface for the editor.
type Manager interface {
	Init(chan<- event.Event, chan<- struct{})
	Open(string) error
	SetSize(int, int)
	Resize(int, int)
	Emit(event.Event)
	State() (map[int]*state.WindowState, layout.Layout, int, error)
	Close()
}
