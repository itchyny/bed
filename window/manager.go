package window

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/itchyny/bed/event"
	"github.com/itchyny/bed/layout"
	"github.com/itchyny/bed/state"
)

// Manager manages the windows and files.
type Manager struct {
	width           int
	height          int
	windows         []*window
	layout          layout.Layout
	mu              *sync.Mutex
	windowIndex     int
	prevWindowIndex int
	prevDir         string
	files           map[string]file
	eventCh         chan<- event.Event
	redrawCh        chan<- struct{}
}

type file struct {
	path string
	file *os.File
	perm os.FileMode
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Init initializes the Manager.
func (m *Manager) Init(eventCh chan<- event.Event, redrawCh chan<- struct{}) {
	m.eventCh, m.redrawCh = eventCh, redrawCh
	m.mu, m.files = new(sync.Mutex), make(map[string]file)
}

// Open a new window.
func (m *Manager) Open(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	window, err := m.open(name)
	if err != nil {
		return err
	}
	return m.init(window)
}

// Read opens a new window from [io.Reader].
func (m *Manager) Read(r io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.read(r)
}

func (m *Manager) init(window *window) error {
	m.addWindow(window)
	m.layout = layout.NewLayout(m.windowIndex).Resize(0, 0, m.width, m.height)
	return nil
}

func (m *Manager) addWindow(window *window) {
	for i, w := range m.windows {
		if w == window {
			m.windowIndex, m.prevWindowIndex = i, m.windowIndex
			return
		}
	}
	m.windows = append(m.windows, window)
	m.windowIndex, m.prevWindowIndex = len(m.windows)-1, m.windowIndex
}

func (m *Manager) open(name string) (*window, error) {
	if name == "" {
		window, err := newWindow(bytes.NewReader(nil), "", "", m.eventCh, m.redrawCh)
		if err != nil {
			return nil, err
		}
		return window, nil
	}
	if name == "#" {
		return m.windows[m.prevWindowIndex], nil
	}
	if strings.HasPrefix(name, "#") {
		index, err := strconv.Atoi(name[1:])
		if err != nil || index <= 0 || len(m.windows) < index {
			return nil, errors.New("invalid window index: " + name)
		}
		return m.windows[index-1], nil
	}
	name, err := expandBacktick(name)
	if err != nil {
		return nil, err
	}
	path, err := expandPath(name)
	if err != nil {
		return nil, err
	}
	r, err := m.openFile(path, name)
	if err != nil {
		return nil, err
	}
	return newWindow(r, path, filepath.Base(path), m.eventCh, m.redrawCh)
}

func (m *Manager) openFile(path, name string) (readAtSeeker, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return bytes.NewReader(nil), nil
	} else if fi.IsDir() {
		return nil, errors.New(name + " is a directory")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	m.addFile(path, f, fi)
	return f, nil
}

func expandBacktick(name string) (string, error) {
	if len(name) <= 2 || name[0] != '`' || name[len(name)-1] != '`' {
		return name, nil
	}
	name = strings.TrimSpace(name[1 : len(name)-1])
	xs := strings.Fields(name)
	if len(xs) < 1 {
		return name, nil
	}
	out, err := exec.Command(xs[0], xs[1:]...).Output()
	if err != nil {
		return name, err
	}
	return strings.TrimSpace(string(out)), nil
}

func expandPath(path string) (string, error) {
	switch {
	case strings.HasPrefix(path, "~"):
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[1:]), nil
	case strings.HasPrefix(path, "$"):
		name, rest, _ := strings.Cut(path[1:], string(filepath.Separator))
		value := os.Getenv(name)
		if value == "" {
			return path, nil
		}
		return filepath.Join(value, rest), nil
	default:
		return filepath.Abs(path)
	}
}

func (m *Manager) read(r io.Reader) error {
	bs, err := func() ([]byte, error) {
		r, stop := newReader(r)
		defer stop()
		return io.ReadAll(r)
	}()
	if err != nil {
		return err
	}
	window, err := newWindow(bytes.NewReader(bs), "", "", m.eventCh, m.redrawCh)
	if err != nil {
		return err
	}
	return m.init(window)
}

type reader struct {
	io.Reader
	abort chan os.Signal
}

func newReader(r io.Reader) (*reader, func()) {
	done := make(chan struct{})
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, os.Interrupt)
	go func() {
		select {
		case <-time.After(time.Second):
			fmt.Fprint(os.Stderr, "Reading stdin took more than 1 second, press <C-c> to abort...")
		case <-done:
		}
	}()
	return &reader{r, abort}, func() {
		signal.Stop(abort)
		close(abort)
		close(done)
	}
}

func (r *reader) Read(p []byte) (int, error) {
	select {
	case <-r.abort:
		return 0, io.EOF
	default:
	}
	return r.Reader.Read(p)
}

// SetSize sets the size of the screen.
func (m *Manager) SetSize(width, height int) {
	m.width, m.height = width, height
}

// Resize sets the size of the screen.
func (m *Manager) Resize(width, height int) {
	if m.width != width || m.height != height {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.width, m.height = width, height
		m.layout = m.layout.Resize(0, 0, width, height)
	}
}

// Emit an event to the current window.
func (m *Manager) Emit(e event.Event) {
	switch e.Type {
	case event.Edit:
		if err := m.edit(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.Enew:
		if err := m.enew(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.New:
		if err := m.newWindow(e, false); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.Vnew:
		if err := m.newWindow(e, true); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.Only:
		if err := m.only(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.Alternative:
		m.alternative(e)
		m.eventCh <- event.Event{Type: event.Redraw}
	case event.Wincmd:
		if e.Arg == "" {
			m.eventCh <- event.Event{Type: event.Error,
				Error: errors.New("an argument is required for " + e.CmdName)}
		} else if err := m.wincmd(e.Arg); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowDown:
		if err := m.wincmd("j"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowUp:
		if err := m.wincmd("k"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowLeft:
		if err := m.wincmd("h"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowRight:
		if err := m.wincmd("l"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowTopLeft:
		if err := m.wincmd("t"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowBottomRight:
		if err := m.wincmd("b"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.FocusWindowPrevious:
		if err := m.wincmd("p"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.MoveWindowTop:
		if err := m.wincmd("K"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.MoveWindowBottom:
		if err := m.wincmd("J"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.MoveWindowLeft:
		if err := m.wincmd("H"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.MoveWindowRight:
		if err := m.wincmd("L"); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Redraw}
		}
	case event.Pwd:
		if e.Arg != "" {
			m.eventCh <- event.Event{Type: event.Error, Error: errors.New("too many arguments for " + e.CmdName)}
			break
		}
		fallthrough
	case event.Chdir:
		if dir, err := m.chdir(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Info, Error: errors.New(dir)}
		}
	case event.Quit:
		if err := m.quit(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		}
	case event.Write:
		if name, n, err := m.write(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else {
			m.eventCh <- event.Event{Type: event.Info,
				Error: fmt.Errorf("%s: %[2]d (0x%[2]x) bytes written", name, n)}
		}
	case event.WriteQuit:
		if _, _, err := m.write(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		} else if err := m.quit(event.Event{}); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		}
	default:
		m.windows[m.windowIndex].emit(e)
	}
}

func (m *Manager) edit(e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := e.Arg
	if name == "" {
		name = m.windows[m.windowIndex].path
	}
	window, err := m.open(name)
	if err != nil {
		return err
	}
	m.addWindow(window)
	m.layout = m.layout.Replace(m.windowIndex)
	return nil
}

func (m *Manager) enew(e event.Event) error {
	if e.Arg != "" {
		return errors.New("too many arguments for " + e.CmdName)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	window, err := m.open("")
	if err != nil {
		return err
	}
	m.addWindow(window)
	m.layout = m.layout.Replace(m.windowIndex)
	return nil
}

func (m *Manager) newWindow(e event.Event, vertical bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	window, err := m.open(e.Arg)
	if err != nil {
		return err
	}
	m.addWindow(window)
	if vertical {
		m.layout = m.layout.SplitLeft(m.windowIndex).Resize(0, 0, m.width, m.height)
	} else {
		m.layout = m.layout.SplitTop(m.windowIndex).Resize(0, 0, m.width, m.height)
	}
	return nil
}

func (m *Manager) only(e event.Event) error {
	if e.Arg != "" {
		return errors.New("too many arguments for " + e.CmdName)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !e.Bang {
		for windowIndex, w := range m.layout.Collect() {
			if window := m.windows[windowIndex]; !w.Active && window.changedTick != window.savedChangedTick {
				return errors.New("you have unsaved changes in " + window.getName() + ", add ! to force :only")
			}
		}
	}
	m.layout = layout.NewLayout(m.windowIndex).Resize(0, 0, m.width, m.height)
	return nil
}

func (m *Manager) alternative(e event.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e.Count == 0 {
		m.windowIndex, m.prevWindowIndex = m.prevWindowIndex, m.windowIndex
	} else if 0 < e.Count && e.Count <= int64(len(m.windows)) {
		m.windowIndex, m.prevWindowIndex = int(e.Count)-1, m.windowIndex
	}
	m.layout = m.layout.Replace(m.windowIndex)
}

func (m *Manager) wincmd(arg string) error {
	switch arg {
	case "n":
		return m.newWindow(event.Event{}, false)
	case "o":
		return m.only(event.Event{})
	case "l":
		m.focus(func(x, y layout.Window) bool {
			return x.LeftMargin()+x.Width()+1 == y.LeftMargin() &&
				y.TopMargin() <= x.TopMargin() &&
				x.TopMargin() < y.TopMargin()+y.Height()
		})
	case "h":
		m.focus(func(x, y layout.Window) bool {
			return y.LeftMargin()+y.Width()+1 == x.LeftMargin() &&
				y.TopMargin() <= x.TopMargin() &&
				x.TopMargin() < y.TopMargin()+y.Height()
		})
	case "k":
		m.focus(func(x, y layout.Window) bool {
			return y.TopMargin()+y.Height() == x.TopMargin() &&
				y.LeftMargin() <= x.LeftMargin() &&
				x.LeftMargin() < y.LeftMargin()+y.Width()
		})
	case "j":
		m.focus(func(x, y layout.Window) bool {
			return x.TopMargin()+x.Height() == y.TopMargin() &&
				y.LeftMargin() <= x.LeftMargin() &&
				x.LeftMargin() < y.LeftMargin()+y.Width()
		})
	case "t":
		m.focus(func(_, y layout.Window) bool {
			return y.LeftMargin() == 0 && y.TopMargin() == 0
		})
	case "b":
		m.focus(func(_, y layout.Window) bool {
			return m.layout.LeftMargin()+m.layout.Width() == y.LeftMargin()+y.Width() &&
				m.layout.TopMargin()+m.layout.Height() == y.TopMargin()+y.Height()
		})
	case "p":
		m.focus(func(_, y layout.Window) bool {
			return y.Index == m.prevWindowIndex
		})
	case "K":
		m.move(func(x layout.Window, y layout.Layout) layout.Layout {
			return layout.Horizontal{Top: x, Bottom: y}
		})
	case "J":
		m.move(func(x layout.Window, y layout.Layout) layout.Layout {
			return layout.Horizontal{Top: y, Bottom: x}
		})
	case "H":
		m.move(func(x layout.Window, y layout.Layout) layout.Layout {
			return layout.Vertical{Left: x, Right: y}
		})
	case "L":
		m.move(func(x layout.Window, y layout.Layout) layout.Layout {
			return layout.Vertical{Left: y, Right: x}
		})
	default:
		return errors.New("Invalid argument for wincmd: " + arg)
	}
	return nil
}

func (m *Manager) focus(search func(layout.Window, layout.Window) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	activeWindow := m.layout.ActiveWindow()
	newWindow := m.layout.Lookup(func(l layout.Window) bool {
		return search(activeWindow, l)
	})
	if newWindow.Index >= 0 {
		m.windowIndex, m.prevWindowIndex = newWindow.Index, m.windowIndex
		m.layout = m.layout.Activate(m.windowIndex)
	}
}

func (m *Manager) move(modifier func(layout.Window, layout.Layout) layout.Layout) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w, h := m.layout.Count()
	if w != 1 || h != 1 {
		activeWindow := m.layout.ActiveWindow()
		m.layout = modifier(activeWindow, m.layout.Close()).Activate(
			activeWindow.Index).Resize(0, 0, m.width, m.height)
	}
}

func (m *Manager) chdir(e event.Event) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e.Arg == "-" && m.prevDir == "" {
		return "", errors.New("no previous working directory")
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if e.Arg == "" {
		return dir, nil
	}
	if e.Arg != "-" {
		dir, m.prevDir = e.Arg, dir
	} else {
		dir, m.prevDir = m.prevDir, dir
	}
	if dir, err = expandPath(dir); err != nil {
		return "", err
	}
	if err = os.Chdir(dir); err != nil {
		return "", err
	}
	return os.Getwd()
}

func (m *Manager) quit(e event.Event) error {
	if e.Arg != "" {
		return errors.New("too many arguments for " + e.CmdName)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	window := m.windows[m.windowIndex]
	if window.changedTick != window.savedChangedTick && !e.Bang {
		return errors.New("you have unsaved changes in " + window.getName() + ", add ! to force :quit")
	}
	w, h := m.layout.Count()
	if w == 1 && h == 1 {
		m.eventCh <- event.Event{Type: event.QuitAll}
	} else {
		m.layout = m.layout.Close().Resize(0, 0, m.width, m.height)
		m.windowIndex, m.prevWindowIndex = m.layout.ActiveWindow().Index, m.windowIndex
		m.eventCh <- event.Event{Type: event.Redraw}
	}
	return nil
}

func (m *Manager) write(e event.Event) (string, int64, error) {
	if e.Range != nil && e.Arg == "" {
		return "", 0, errors.New("cannot overwrite partially with " + e.CmdName)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	window := m.windows[m.windowIndex]
	var path string
	name := e.Arg
	if name == "" {
		if window.name == "" {
			return "", 0, errors.New("no file name")
		}
		path, name = window.path, window.name
	} else {
		var err error
		path, err = expandPath(name)
		if err != nil {
			return "", 0, err
		}
	}
	if runtime.GOOS == "windows" && m.opened(path) {
		return "", 0, errors.New("cannot overwrite the original file on Windows")
	}
	if window.path == "" && window.name == "" {
		window.setPathName(path, filepath.Base(path))
	}
	tmpf, err := os.OpenFile(
		path+"-"+strconv.FormatUint(rand.Uint64(), 16),
		os.O_RDWR|os.O_CREATE|os.O_EXCL, m.filePerm(path),
	) //#nosec G404
	if err != nil {
		return "", 0, err
	}
	defer os.Remove(tmpf.Name())
	n, err := window.writeTo(e.Range, tmpf)
	tmpf.Close()
	if err != nil {
		return "", 0, err
	}
	if window.path == path {
		window.savedChangedTick = window.changedTick
	}
	return name, n, os.Rename(tmpf.Name(), path)
}

func (m *Manager) addFile(path string, f *os.File, fi os.FileInfo) {
	m.files[path] = file{path: path, file: f, perm: fi.Mode().Perm()}
}

func (m *Manager) opened(path string) bool {
	_, ok := m.files[path]
	return ok
}

func (m *Manager) filePerm(path string) os.FileMode {
	if f, ok := m.files[path]; ok {
		return f.perm
	}
	return os.FileMode(0o644)
}

// State returns the state of the windows.
func (m *Manager) State() (map[int]*state.WindowState, layout.Layout, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	layouts := m.layout.Collect()
	states := make(map[int]*state.WindowState, len(m.windows))
	for i, window := range m.windows {
		if l, ok := layouts[i]; ok {
			var err error
			if states[i], err = window.state(
				hexWindowWidth(l.Width()), max(l.Height()-2, 1),
			); err != nil {
				return nil, m.layout, 0, err
			}
		}
	}
	return states, m.layout, m.windowIndex, nil
}

func hexWindowWidth(width int) int {
	width = min(max((width-18)/4, 4), 256)
	return width & (0b11 << (bits.Len(uint(width)) - 2))
}

// Close the Manager.
func (m *Manager) Close() {
	for _, f := range m.files {
		f.file.Close()
	}
}
