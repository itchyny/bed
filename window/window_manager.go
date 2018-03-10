package window

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/itchyny/bed/core"
)

// WindowManager manages the windows and files.
type WindowManager struct {
	height   int64
	window   *Window
	files    []file
	redrawCh chan<- struct{}
}

type file struct {
	name string
	file *os.File
	perm os.FileMode
}

// NewWindowManager creates a new WindowManager.
func NewWindowManager() *WindowManager {
	return &WindowManager{}
}

func (wm *WindowManager) Init(redrawCh chan<- struct{}) error {
	wm.redrawCh = redrawCh
	return nil
}

func (wm *WindowManager) Open(filename string) (err error) {
	if filename == "" {
		if wm.window, err = NewWindow(bytes.NewReader(nil), "", "", wm.height, 16, wm.redrawCh); err != nil {
			return err
		}
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if wm.window, err = NewWindow(bytes.NewReader(nil), filename, filepath.Base(filename), wm.height, 16, wm.redrawCh); err != nil {
			return err
		}
		return nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	wm.files = append(wm.files, file{name: filename, file: f, perm: info.Mode().Perm()})
	if wm.window, err = NewWindow(f, filename, filepath.Base(filename), wm.height, 16, wm.redrawCh); err != nil {
		return err
	}
	return nil
}

func (wm *WindowManager) SetHeight(height int) {
	wm.height = int64(height)
}

func (wm *WindowManager) Run() {
	wm.window.Run()
}

func (wm *WindowManager) Emit(event Event) {
	wm.window.eventCh <- event
}

func (wm *WindowManager) State() (State, error) {
	return wm.window.State()
}

func (wm *WindowManager) WriteFile(name string) error {
	perm := os.FileMode(0644)
	if name == "" {
		name = wm.window.filename
	}
	if name == "" {
		return errors.New("no file name")
	}
	if wm.window.filename == "" {
		wm.window.filename = name
	}
	for _, f := range wm.files {
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
	wm.window.buffer.Seek(0, io.SeekStart)
	_, err = io.Copy(tmpf, wm.window.buffer)
	tmpf.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpf.Name(), name)
}

func (wm *WindowManager) Close() {
	for _, f := range wm.files {
		f.file.Close()
	}
	close(wm.window.eventCh)
}
