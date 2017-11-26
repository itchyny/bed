package tui

import (
	"github.com/itchyny/bed/core"
	"github.com/nsf/termbox-go"
)

func eventToKey(event termbox.Event) core.Key {
	if key, ok := keyMap[event.Key]; ok {
		return key
	}
	return core.NewKey(string(event.Ch))
}

var keyMap = map[termbox.Key]core.Key{
	termbox.KeyF1:  core.Key{Key: "f1"},
	termbox.KeyF2:  core.Key{Key: "f2"},
	termbox.KeyF3:  core.Key{Key: "f3"},
	termbox.KeyF4:  core.Key{Key: "f4"},
	termbox.KeyF5:  core.Key{Key: "f5"},
	termbox.KeyF6:  core.Key{Key: "f6"},
	termbox.KeyF7:  core.Key{Key: "f7"},
	termbox.KeyF8:  core.Key{Key: "f8"},
	termbox.KeyF9:  core.Key{Key: "f9"},
	termbox.KeyF10: core.Key{Key: "f10"},
	termbox.KeyF11: core.Key{Key: "f11"},
	termbox.KeyF12: core.Key{Key: "f12"},

	termbox.KeyInsert: core.Key{Key: "insert"},
	termbox.KeyDelete: core.Key{Key: "delete"},
	termbox.KeyHome:   core.Key{Key: "home"},
	termbox.KeyEnd:    core.Key{Key: "end"},
	termbox.KeyPgup:   core.Key{Key: "pgup"},
	termbox.KeyPgdn:   core.Key{Key: "pgdn"},

	termbox.KeyArrowUp:    core.Key{Key: "up"},
	termbox.KeyArrowDown:  core.Key{Key: "down"},
	termbox.KeyArrowLeft:  core.Key{Key: "left"},
	termbox.KeyArrowRight: core.Key{Key: "right"},

	termbox.KeyCtrlA:     core.Key{Key: "a", Ctrl: true},
	termbox.KeyCtrlB:     core.Key{Key: "b", Ctrl: true},
	termbox.KeyCtrlC:     core.Key{Key: "c", Ctrl: true},
	termbox.KeyCtrlD:     core.Key{Key: "d", Ctrl: true},
	termbox.KeyCtrlE:     core.Key{Key: "e", Ctrl: true},
	termbox.KeyCtrlF:     core.Key{Key: "f", Ctrl: true},
	termbox.KeyCtrlG:     core.Key{Key: "g", Ctrl: true},
	termbox.KeyBackspace: core.Key{Key: "backspace"},
	termbox.KeyTab:       core.Key{Key: "tab"},
	termbox.KeyCtrlJ:     core.Key{Key: "j", Ctrl: true},
	termbox.KeyCtrlK:     core.Key{Key: "k", Ctrl: true},
	termbox.KeyCtrlL:     core.Key{Key: "l", Ctrl: true},
	termbox.KeyEnter:     core.Key{Key: "enter"},
	termbox.KeyCtrlN:     core.Key{Key: "n", Ctrl: true},
	termbox.KeyCtrlO:     core.Key{Key: "o", Ctrl: true},
	termbox.KeyCtrlP:     core.Key{Key: "p", Ctrl: true},
	termbox.KeyCtrlQ:     core.Key{Key: "q", Ctrl: true},
	termbox.KeyCtrlR:     core.Key{Key: "r", Ctrl: true},
	termbox.KeyCtrlS:     core.Key{Key: "s", Ctrl: true},
	termbox.KeyCtrlT:     core.Key{Key: "t", Ctrl: true},
	termbox.KeyCtrlU:     core.Key{Key: "u", Ctrl: true},
	termbox.KeyCtrlV:     core.Key{Key: "v", Ctrl: true},
	termbox.KeyCtrlW:     core.Key{Key: "w", Ctrl: true},
	termbox.KeyCtrlX:     core.Key{Key: "x", Ctrl: true},
	termbox.KeyCtrlY:     core.Key{Key: "y", Ctrl: true},
	termbox.KeyCtrlZ:     core.Key{Key: "z", Ctrl: true},
	termbox.KeyEsc:       core.Key{Key: "esc"},
	termbox.KeyCtrl4:     core.Key{Key: "4", Ctrl: true},
	termbox.KeyCtrl5:     core.Key{Key: "5", Ctrl: true},
	termbox.KeyCtrl6:     core.Key{Key: "6", Ctrl: true},
	termbox.KeyCtrlSlash: core.Key{Key: "slash", Ctrl: true},
	termbox.KeySpace:     core.Key{Key: "space"},
	termbox.KeyCtrl8:     core.Key{Key: "8", Ctrl: true},
}
