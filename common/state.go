package common

// State holds the state of the editor to display the user interface.
type State struct {
	Name          string
	Width         int
	Offset        int64
	Cursor        int64
	Bytes         []byte
	Size          int
	Length        int64
	Mode          Mode
	Pending       bool
	PendingByte   byte
	EditedIndices []int64
	FocusText     bool
	Cmdline       []rune
	CmdlineCursor int
	Error         error
	ErrorType     int
}

// Message types
const (
	MessageInfo = iota
	MessageError
)
