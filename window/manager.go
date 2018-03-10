package window

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/itchyny/bed/core"
)

// Manager manages the windows and files.
type Manager struct {
	height   int64
	window   *window
	files    []file
	eventCh  chan<- Event
	redrawCh chan<- struct{}
}

type file struct {
	name string
	file *os.File
	perm os.FileMode
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Init initializes the Manager.
func (m *Manager) Init(eventCh chan<- Event, redrawCh chan<- struct{}) error {
	m.eventCh, m.redrawCh = eventCh, redrawCh
	return nil
}

// Open a new window.
func (m *Manager) Open(filename string) (err error) {
	if filename == "" {
		if m.window, err = newWindow(bytes.NewReader(nil), "", "", m.height, 16, m.redrawCh); err != nil {
			return err
		}
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if m.window, err = newWindow(bytes.NewReader(nil), filename, filepath.Base(filename), m.height, 16, m.redrawCh); err != nil {
			return err
		}
		return nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	m.files = append(m.files, file{name: filename, file: f, perm: info.Mode().Perm()})
	if m.window, err = newWindow(f, filename, filepath.Base(filename), m.height, 16, m.redrawCh); err != nil {
		return err
	}
	return nil
}

// SetHeight sets the height.
func (m *Manager) SetHeight(height int) {
	m.height = int64(height)
}

// Run the Manager.
func (m *Manager) Run() {
	m.window.Run()
}

// Emit an event to the current window.
func (m *Manager) Emit(event Event) {
	switch event.Type {
	case EventEdit:
		if len(event.Args) > 1 {
			m.eventCh <- Event{Type: EventError, Error: fmt.Errorf("too many arguments for %s", event.CmdName)}
		} else if len(event.Args) == 0 {
			m.eventCh <- Event{Type: EventError, Error: errors.New("no file name")}
		} else {
			if err := m.Open(event.Args[0]); err != nil {
				m.eventCh <- Event{Type: EventError, Error: err}
			}
			go m.Run()
			m.eventCh <- Event{Type: EventError, Error: nil}
		}
	case EventWrite:
		if len(event.Args) > 1 {
			m.eventCh <- Event{Type: EventError, Error: fmt.Errorf("too many arguments for %s", event.CmdName)}
		} else {
			var name string
			if len(event.Args) > 0 {
				name = event.Args[0]
			}
			if err := m.writeFile(name); err != nil {
				m.eventCh <- Event{Type: EventError, Error: err}
			}
		}
	case EventWriteQuit:
		if len(event.Args) > 0 {
			m.eventCh <- Event{Type: EventError, Error: fmt.Errorf("too many arguments for %s", event.CmdName)}
		} else {
			if err := m.writeFile(""); err != nil {
				m.eventCh <- Event{Type: EventError, Error: err}
			} else {
				m.eventCh <- Event{Type: EventQuit}
			}
		}
	default:
		m.window.eventCh <- event
	}
}

// State returns the state of the windows.
func (m *Manager) State() (State, error) {
	return m.window.State()
}

func (m *Manager) writeFile(name string) error {
	perm := os.FileMode(0644)
	if name == "" {
		name = m.window.filename
	}
	if name == "" {
		return errors.New("no file name")
	}
	if m.window.filename == "" {
		m.window.filename = name
	}
	for _, f := range m.files {
		if f.name == name {
			perm = f.perm
		}
	}
	tmpf, err := os.OpenFile(
		name+"-"+strconv.FormatUint(rand.Uint64(), 16), os.O_RDWR|os.O_CREATE|os.O_EXCL, perm,
	)
	if err != nil {
		return err
	}
	defer os.Remove(tmpf.Name())
	m.window.buffer.Seek(0, io.SeekStart)
	_, err = io.Copy(tmpf, m.window.buffer)
	tmpf.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpf.Name(), name)
}

// Close the Manager.
func (m *Manager) Close() {
	for _, f := range m.files {
		f.file.Close()
	}
	close(m.window.eventCh)
}
