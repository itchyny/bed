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
	files           []file
	eventCh         chan<- event.Event
	redrawCh        chan<- struct{}
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
func (m *Manager) Init(eventCh chan<- event.Event, redrawCh chan<- struct{}) {
	m.eventCh, m.redrawCh = eventCh, redrawCh
	m.mu = new(sync.Mutex)
}

// Open a new window.
func (m *Manager) Open(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	window, err := m.open(filename)
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

func (m *Manager) open(filename string) (*window, error) {
	if filename == "" {
		window, err := newWindow(bytes.NewReader(nil), "", "", m.eventCh, m.redrawCh)
		if err != nil {
			return nil, err
		}
		return window, nil
	}
	if filename == "#" {
		return m.windows[m.prevWindowIndex], nil
	}
	if strings.HasPrefix(filename, "#") {
		index, err := strconv.Atoi(filename[1:])
		if err != nil || index <= 0 || len(m.windows) < index {
			return nil, errors.New("invalid window index: " + filename)
		}
		return m.windows[index-1], nil
	}
	name, err := expandBacktick(filename)
	if err != nil {
		return nil, err
	}
	name, err = expandHomedir(name)
	if err != nil {
		return nil, err
	}
	filename = name
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		window, err := newWindow(bytes.NewReader(nil), filename, filepath.Base(filename), m.eventCh, m.redrawCh)
		if err != nil {
			return nil, err
		}
		return window, nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New(filename + " is a directory")
	}
	m.files = append(m.files, file{name: filename, file: f, perm: info.Mode().Perm()})
	window, err := newWindow(f, filename, filepath.Base(filename), m.eventCh, m.redrawCh)
	if err != nil {
		return nil, err
	}
	return window, nil
}

func expandBacktick(filename string) (string, error) {
	if !strings.HasPrefix(filename, "`") ||
		!strings.HasSuffix(filename, "`") || len(filename) <= 2 {
		return filename, nil
	}
	filename = strings.TrimSpace(filename[1 : len(filename)-1])
	xs := strings.Fields(filename)
	if len(xs) < 1 {
		return filename, nil
	}
	out, err := exec.Command(xs[0], xs[1:]...).Output()
	if err != nil {
		return filename, err
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) read(r io.Reader) error {
	done := make(chan struct{})
	defer close(done)
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, os.Interrupt)
	defer signal.Stop(abort)
	go func() {
		select {
		case <-time.After(time.Second):
			fmt.Fprint(os.Stderr, "Reading stdin took more than 1 second, press <C-c> to abort...")
		case <-done:
		}
	}()
	bs, err := func() ([]byte, error) {
		// ref: io.ReadAll
		bs := make([]byte, 0, 1024)
		for {
			n, err := r.Read(bs[len(bs):cap(bs)])
			bs = bs[:len(bs)+n]
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				return bs, err
			}
			select {
			case <-time.After(10 * time.Millisecond):
			case <-abort:
				return bs, nil
			}
			if len(bs) == cap(bs) {
				bs = append(bs, 0)[:len(bs)]
			}
		}
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
		if len(e.Arg) == 0 {
			m.eventCh <- event.Event{Type: event.Error, Error: errors.New("an argument is required for " + e.CmdName)}
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
	case event.Quit:
		if err := m.quit(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		}
	case event.Write:
		if err := m.write(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		}
	case event.WriteQuit:
		if err := m.writeQuit(e); err != nil {
			m.eventCh <- event.Event{Type: event.Error, Error: err}
		}
	default:
		m.windows[m.windowIndex].emit(e)
	}
}

func (m *Manager) edit(e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var name string
	if len(e.Arg) == 0 {
		name = m.windows[m.windowIndex].filename
	} else {
		name = e.Arg
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
	if len(e.Arg) > 0 {
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
	if len(e.Arg) > 0 {
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

func (m *Manager) quit(e event.Event) error {
	if len(e.Arg) > 0 {
		return errors.New("too many arguments for " + e.CmdName)
	}
	window := m.windows[m.windowIndex]
	if window.changedTick != window.savedChangedTick && !e.Bang {
		return errors.New("you have unsaved changes in " + window.getName() + ", add ! to force :quit")
	}
	w, h := m.layout.Count()
	if w == 1 && h == 1 {
		m.eventCh <- event.Event{Type: event.QuitAll}
	} else {
		m.mu.Lock()
		m.layout = m.layout.Close().Resize(0, 0, m.width, m.height)
		m.windowIndex, m.prevWindowIndex = m.layout.ActiveWindow().Index, m.windowIndex
		m.mu.Unlock()
		m.eventCh <- event.Event{Type: event.Redraw}
	}
	return nil
}

func (m *Manager) write(e event.Event) error {
	if e.Range != nil && e.Arg == "" {
		return errors.New("cannot overwrite partially with " + e.CmdName)
	}
	filename, n, err := m.writeFile(e.Range, e.Arg)
	if err != nil {
		return err
	}
	m.eventCh <- event.Event{Type: event.Info, Error: fmt.Errorf("%s: %d (0x%x) bytes written", filename, n, n)}
	return nil
}

func (m *Manager) writeQuit(e event.Event) error {
	if len(e.Arg) > 0 {
		return errors.New("too many arguments for " + e.CmdName)
	}
	if e.Range != nil {
		return errors.New("range not allowed for " + e.CmdName)
	}
	if _, _, err := m.writeFile(nil, ""); err != nil {
		return err
	}
	return m.quit(e)
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

func (m *Manager) writeFile(r *event.Range, name string) (string, int64, error) {
	window := m.windows[m.windowIndex]
	if name == "" {
		name = window.filename
	}
	if name == "" {
		return name, 0, errors.New("no file name")
	}
	if runtime.GOOS == "windows" && m.opened(name) {
		return name, 0, errors.New("cannot overwrite the original file on Windows")
	}
	var err error
	if name, err = expandHomedir(name); err != nil {
		return name, 0, err
	}
	if window.filename == "" && window.name == "" {
		window.mu.Lock()
		window.filename = name
		window.name = filepath.Base(name)
		window.mu.Unlock()
	}
	tmpf, err := os.OpenFile(
		name+"-"+strconv.FormatUint(rand.Uint64(), 16),
		os.O_RDWR|os.O_CREATE|os.O_EXCL, m.filePerm(name),
	) //#nosec G404
	if err != nil {
		return name, 0, err
	}
	defer os.Remove(tmpf.Name())
	n, err := window.writeTo(r, tmpf)
	tmpf.Close()
	if err != nil {
		return name, 0, err
	}
	if window.filename == name {
		window.savedChangedTick = window.changedTick
	}
	return name, n, os.Rename(tmpf.Name(), name)
}

func (m *Manager) filePerm(name string) os.FileMode {
	for _, f := range m.files {
		if f.name == name {
			return f.perm // keep the permission of the original file
		}
	}
	return os.FileMode(0o644)
}

func (m *Manager) opened(name string) bool {
	for _, f := range m.files {
		if f.name == name {
			return true
		}
	}
	return false
}

// Close the Manager.
func (m *Manager) Close() {
	for _, f := range m.files {
		f.file.Close()
	}
}

func expandHomedir(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, path[1:]), nil
}
