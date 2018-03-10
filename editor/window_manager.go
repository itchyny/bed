package editor

import . "github.com/itchyny/bed/core"

// WindowManager defines the required window manager interface for the editor.
type WindowManager interface {
	Init(chan<- Event, chan<- struct{}) error
	Open(string) error
	SetHeight(int)
	Run()
	Emit(Event)
	State() (State, error)
	Close()
}
