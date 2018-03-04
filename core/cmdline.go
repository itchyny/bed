package core

// Cmdline defines the required cmdline interface for the editor.
type Cmdline interface {
	CursorLeft()
	CursorRight()
	CursorHead()
	CursorEnd()
	Backspace()
	Delete()
	Clear()
	ClearToHead()
	Insert(rune)
	Get() ([]rune, int)
}
