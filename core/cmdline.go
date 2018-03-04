package core

// Cmdline defines the required cmdline interface for the editor.
type Cmdline interface {
	Init(ch chan<- Event) error
	CursorLeft()
	CursorRight()
	CursorHead()
	CursorEnd()
	Backspace()
	Delete()
	DeleteWord()
	Clear()
	ClearToHead()
	Insert(rune)
	Get() ([]rune, int)
	Execute()
}
