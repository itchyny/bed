package mode

// Mode ...
type Mode int

// Modes
const (
	Normal Mode = iota
	Insert
	Replace
	Cmdline
	Search
)
