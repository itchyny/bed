package core

import (
	"os"
	"path/filepath"
)

// File holds file descriptor and file name.
type File struct {
	*os.File
	name     string
	basename string
}

// NewFile creates a new File.
func NewFile(filename string) (*File, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &File{
		File:     fd,
		name:     filename,
		basename: filepath.Base(filename),
	}, nil
}
