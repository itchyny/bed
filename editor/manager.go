package editor

import (
	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/layout"
)

// Manager defines the required window manager interface for the editor.
type Manager interface {
	Init(chan<- Event, chan<- struct{})
	Open(string) error
	SetSize(int, int)
	Resize(int, int)
	Run()
	Emit(Event)
	State() (map[int]*WindowState, layout.Layout, int, error)
	Close()
}
