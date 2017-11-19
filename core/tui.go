package core

import (
	termbox "github.com/nsf/termbox-go"
)

// Tui implements UI
type Tui struct {
	width  int
	height int
}

func NewTui() *Tui {
	return &Tui{}
}

func (ui *Tui) Init() error {
	return termbox.Init()
}

func (ui *Tui) Start(cb func(int, int) error) error {
	ui.width, ui.height = termbox.Size()
	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
	if err := cb(ui.height, ui.width); err != nil {
		return err
	}
loop:
	for {
		select {
		case event := <-events:
			if event.Type == termbox.EventKey {
				if event.Ch == 'q' || event.Key == termbox.KeyCtrlC || event.Key == termbox.KeyCtrlD {
					break loop
				}
			}
		}
	}
	return nil
}

func (ui *Tui) SetLine(line int, str string) error {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	for i, c := range str {
		termbox.SetCell(i, line, c, fg, bg)
	}
	return termbox.Flush()
}

func (ui *Tui) Close() error {
	termbox.Close()
	return nil
}
