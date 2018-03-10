package editor

import . "github.com/itchyny/bed/core"

// WindowManager defines the required window manager interface for the editor.
type WindowManager interface {
	Init(chan<- struct{}) error
	Open(string) error
	SetHeight(height int)
	Run()
	Emit(event Event)
	State() (State, error)
	WriteFile(name string) error
	Close()
}
