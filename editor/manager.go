package editor

import . "github.com/itchyny/bed/common"

// Manager defines the required window manager interface for the editor.
type Manager interface {
	Init(chan<- Event, chan<- struct{}) error
	Open(string) error
	SetSize(int, int)
	Resize(int, int)
	Run()
	Emit(Event)
	State() (map[int]*WindowState, Layout, int, error)
	Close()
}
