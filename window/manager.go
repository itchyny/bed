package window

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/mitchellh/go-homedir"

	. "github.com/itchyny/bed/common"
	"github.com/itchyny/bed/mathutil"
)

// Manager manages the windows and files.
type Manager struct {
	width           int
	height          int
	windows         []*window
	layout          Layout
	mu              *sync.Mutex
	windowIndex     int
	prevWindowIndex int
	files           []file
	eventCh         chan<- Event
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
func (m *Manager) Init(eventCh chan<- Event, redrawCh chan<- struct{}) {
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
	m.windows = append(m.windows, window)
	m.windowIndex, m.prevWindowIndex = len(m.windows)-1, m.windowIndex
	m.layout = NewLayout(m.windowIndex).Resize(0, 0, m.width, m.height)
	return nil
}

func (m *Manager) open(filename string) (*window, error) {
	if filename == "" {
		window, err := newWindow(bytes.NewReader(nil), "", "", m.redrawCh)
		if err != nil {
			return nil, err
		}
		return window, nil
	}
	name, err := homedir.Expand(filename)
	if err != nil {
		return nil, err
	}
	filename = name
	f, err := os.Open(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		window, err := newWindow(bytes.NewReader(nil), filename, filepath.Base(filename), m.redrawCh)
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
		return nil, fmt.Errorf("%s is a directory", filename)
	}
	m.files = append(m.files, file{name: filename, file: f, perm: info.Mode().Perm()})
	window, err := newWindow(f, filename, filepath.Base(filename), m.redrawCh)
	if err != nil {
		return nil, err
	}
	return window, nil
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

// Run the Manager.
func (m *Manager) Run() {
	m.windows[m.windowIndex].Run()
}

// Emit an event to the current window.
func (m *Manager) Emit(event Event) {
	switch event.Type {
	case EventCursorGotoAbs:
		if err := m.cursorGoto(event); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		}
	case EventCursorGotoRel:
		if err := m.cursorGoto(event); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		}
	case EventEdit:
		if err := m.edit(event); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		} else {
			m.eventCh <- Event{Type: EventRedraw}
		}
	case EventNew:
		if err := m.newWindow(event, false); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		} else {
			m.eventCh <- Event{Type: EventRedraw}
		}
	case EventVnew:
		if err := m.newWindow(event, true); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		} else {
			m.eventCh <- Event{Type: EventRedraw}
		}
	case EventWincmd:
		if len(event.Arg) == 0 {
			m.eventCh <- Event{Type: EventError, Error: fmt.Errorf("an argument is required for %s", event.CmdName)}
		} else if err := m.wincmd(event.Arg); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		} else {
			m.eventCh <- Event{Type: EventRedraw}
		}
	case EventFocusWindowDown:
		m.wincmd("j")
		m.eventCh <- Event{Type: EventRedraw}
	case EventFocusWindowUp:
		m.wincmd("k")
		m.eventCh <- Event{Type: EventRedraw}
	case EventFocusWindowLeft:
		m.wincmd("h")
		m.eventCh <- Event{Type: EventRedraw}
	case EventFocusWindowRight:
		m.wincmd("l")
		m.eventCh <- Event{Type: EventRedraw}
	case EventFocusWindowTopLeft:
		m.wincmd("t")
		m.eventCh <- Event{Type: EventRedraw}
	case EventFocusWindowBottomRight:
		m.wincmd("b")
		m.eventCh <- Event{Type: EventRedraw}
	case EventFocusWindowPrevious:
		m.wincmd("p")
		m.eventCh <- Event{Type: EventRedraw}
	case EventMoveWindowTop:
		m.wincmd("K")
		m.eventCh <- Event{Type: EventRedraw}
	case EventMoveWindowBottom:
		m.wincmd("J")
		m.eventCh <- Event{Type: EventRedraw}
	case EventMoveWindowLeft:
		m.wincmd("H")
		m.eventCh <- Event{Type: EventRedraw}
	case EventMoveWindowRight:
		m.wincmd("L")
		m.eventCh <- Event{Type: EventRedraw}
	case EventQuit:
		if err := m.quit(event); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		}
	case EventWrite:
		if err := m.write(event); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		}
	case EventWriteQuit:
		if err := m.writeQuit(event); err != nil {
			m.eventCh <- Event{Type: EventError, Error: err}
		}
	default:
		m.windows[m.windowIndex].eventCh <- event
	}
}

func (m *Manager) cursorGoto(event Event) error {
	event.Count = parseGotoPos(event.Arg)
	m.windows[m.windowIndex].eventCh <- event
	return nil
}

func parseGotoPos(pos string) int64 {
	switch pos {
	case "$":
		return math.MaxInt64
	case "+":
		return 1
	case "-":
		return -1
	}
	count, sign := int64(0), int64(1)
	for _, c := range pos {
		count *= 0x10
		if '0' <= c && c <= '9' {
			count += int64(c - '0')
		} else if 'a' <= c && c <= 'f' {
			count += int64(c - 'a' + 0x0a)
		} else if c == '-' {
			sign = -1
		}
	}
	return sign * count
}

func (m *Manager) edit(event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var name string
	if len(event.Arg) == 0 {
		name = m.windows[m.windowIndex].filename
	} else {
		name = event.Arg
	}
	window, err := m.open(name)
	if err != nil {
		return err
	}
	m.windows = append(m.windows, window)
	m.windowIndex, m.prevWindowIndex = len(m.windows)-1, m.windowIndex
	m.layout = m.layout.Replace(m.windowIndex)
	go m.Run()
	return nil
}

func (m *Manager) newWindow(event Event, vertical bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	window, err := m.open(event.Arg)
	if err != nil {
		return err
	}
	m.windows = append(m.windows, window)
	m.windowIndex, m.prevWindowIndex = len(m.windows)-1, m.windowIndex
	if vertical {
		m.layout = m.layout.SplitLeft(m.windowIndex).Resize(0, 0, m.width, m.height)
	} else {
		m.layout = m.layout.SplitTop(m.windowIndex).Resize(0, 0, m.width, m.height)
	}
	go m.Run()
	return nil
}

func (m *Manager) wincmd(arg string) error {
	switch arg {
	case "n":
		return m.newWindow(Event{}, false)
	case "l":
		m.focus(func(x, y LayoutWindow) bool {
			return x.LeftMargin()+x.Width()+1 == y.LeftMargin() &&
				y.TopMargin() <= x.TopMargin() &&
				x.TopMargin() < y.TopMargin()+y.Height()
		})
	case "h":
		m.focus(func(x, y LayoutWindow) bool {
			return y.LeftMargin()+y.Width()+1 == x.LeftMargin() &&
				y.TopMargin() <= x.TopMargin() &&
				x.TopMargin() < y.TopMargin()+y.Height()
		})
	case "k":
		m.focus(func(x, y LayoutWindow) bool {
			return y.TopMargin()+y.Height() == x.TopMargin() &&
				y.LeftMargin() <= x.LeftMargin() &&
				x.LeftMargin() < y.LeftMargin()+y.Width()
		})
	case "j":
		m.focus(func(x, y LayoutWindow) bool {
			return x.TopMargin()+x.Height() == y.TopMargin() &&
				y.LeftMargin() <= x.LeftMargin() &&
				x.LeftMargin() < y.LeftMargin()+y.Width()
		})
	case "t":
		m.focus(func(_, y LayoutWindow) bool {
			return y.LeftMargin() == 0 && y.TopMargin() == 0
		})
	case "b":
		m.focus(func(_, y LayoutWindow) bool {
			return m.layout.LeftMargin()+m.layout.Width() == y.LeftMargin()+y.Width() &&
				m.layout.TopMargin()+m.layout.Height() == y.TopMargin()+y.Height()
		})
	case "p":
		m.focus(func(_, y LayoutWindow) bool {
			return y.Index == m.prevWindowIndex
		})
	case "K":
		m.move(func(x LayoutWindow, y Layout) Layout {
			return LayoutHorizontal{Top: x, Bottom: y}
		})
	case "J":
		m.move(func(x LayoutWindow, y Layout) Layout {
			return LayoutHorizontal{Top: y, Bottom: x}
		})
	case "H":
		m.move(func(x LayoutWindow, y Layout) Layout {
			return LayoutVertical{Left: x, Right: y}
		})
	case "L":
		m.move(func(x LayoutWindow, y Layout) Layout {
			return LayoutVertical{Left: y, Right: x}
		})
	}
	// TODO: return error
	return nil
}

func (m *Manager) focus(search func(LayoutWindow, LayoutWindow) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	activeWindow := m.layout.ActiveWindow()
	newWindow := m.layout.Lookup(func(l LayoutWindow) bool {
		return search(activeWindow, l)
	})
	if newWindow.Index >= 0 {
		m.windowIndex, m.prevWindowIndex = newWindow.Index, m.windowIndex
		m.layout = m.layout.Activate(m.windowIndex)
	}
}

func (m *Manager) move(modifier func(LayoutWindow, Layout) Layout) {
	m.mu.Lock()
	defer m.mu.Unlock()
	activeWindow := m.layout.ActiveWindow()
	m.layout = modifier(activeWindow, m.layout.Close()).Activate(
		activeWindow.Index).Resize(0, 0, m.width, m.height)
}

func (m *Manager) quit(event Event) error {
	if len(event.Arg) > 0 {
		return fmt.Errorf("too many arguments for %s", event.CmdName)
	}
	w, h := m.layout.Count()
	if w == 1 && h == 1 {
		m.eventCh <- Event{Type: EventQuitAll}
	} else {
		m.mu.Lock()
		m.layout = m.layout.Close().Resize(0, 0, m.width, m.height)
		m.windowIndex, m.prevWindowIndex = m.layout.ActiveWindow().Index, m.windowIndex
		m.mu.Unlock()
		m.eventCh <- Event{Type: EventRedraw}
	}
	return nil
}

func (m *Manager) write(event Event) error {
	filename, n, err := m.writeFile(event.Arg)
	if err != nil {
		return err
	}
	m.eventCh <- Event{Type: EventInfo, Error: fmt.Errorf("%s: %d (0x%x) bytes written", filename, n, n)}
	return nil
}

func (m *Manager) writeQuit(event Event) error {
	if len(event.Arg) > 0 {
		return fmt.Errorf("too many arguments for %s", event.CmdName)
	}
	if _, _, err := m.writeFile(""); err != nil {
		return err
	}
	m.eventCh <- Event{Type: EventQuit}
	return nil
}

// State returns the state of the windows.
func (m *Manager) State() (map[int]*WindowState, Layout, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	layouts := m.layout.Collect()
	states := make(map[int]*WindowState, len(m.windows))
	for i, window := range m.windows {
		if l, ok := layouts[i]; ok {
			window.setSize(hexWindowWidth(l.Width()), mathutil.MaxInt(l.Height()-2, 1))
			var err error
			if states[i], err = window.State(); err != nil {
				return nil, m.layout, 0, err
			}
		}
	}
	return states, m.layout, m.windowIndex, nil
}

func hexWindowWidth(width int) int {
	if width > 146 {
		return 32
	} else if width > 82 {
		return 16
	} else if width > 50 {
		return 8
	}
	return 4
}

func (m *Manager) writeFile(name string) (string, int64, error) {
	window := m.windows[m.windowIndex]
	perm := os.FileMode(0644)
	if name == "" {
		name = window.filename
	}
	if name == "" {
		return name, 0, errors.New("no file name")
	}
	if window.filename == "" {
		window.filename = name
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
		return name, 0, err
	}
	defer os.Remove(tmpf.Name())
	n, err := window.writeTo(tmpf)
	tmpf.Close()
	if err != nil {
		return name, 0, err
	}
	return name, n, os.Rename(tmpf.Name(), name)
}

// Close the Manager.
func (m *Manager) Close() {
	for _, f := range m.files {
		f.file.Close()
	}
	for _, w := range m.windows {
		w.Close()
	}
}
