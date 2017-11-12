package core

import (
	"os"
)

type Editor struct {
	buffer *Buffer
}

func NewEditor() *Editor {
	return &Editor{}
}

func (e *Editor) Open(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	e.buffer = NewBuffer(file)
	return nil
}
