package editor

import . "github.com/itchyny/bed/common"

// Manager defines the required window manager interface for the editor.
type Manager interface {
	Init(chan<- Event, chan<- struct{}) error
	Open(string) error
	SetHeight(int)
	Run()
	Emit(Event)
	State() ([]WindowState, error)
	Close()
}
