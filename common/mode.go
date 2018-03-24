package common

// Mode ...
type Mode int

// Modes
const (
	ModeNormal = iota
	ModeInsert
	ModeReplace
	ModeCmdline
	ModeSearch
)
