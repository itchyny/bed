package tui

import (
	"github.com/gdamore/tcell"

	"github.com/itchyny/bed/key"
)

func eventToKey(event *tcell.EventKey) key.Key {
	if key, ok := keyMap[event.Key()]; ok {
		return key
	}
	return key.Key(event.Rune())
}

var keyMap = map[tcell.Key]key.Key{
	tcell.KeyF1:  key.Key("f1"),
	tcell.KeyF2:  key.Key("f2"),
	tcell.KeyF3:  key.Key("f3"),
	tcell.KeyF4:  key.Key("f4"),
	tcell.KeyF5:  key.Key("f5"),
	tcell.KeyF6:  key.Key("f6"),
	tcell.KeyF7:  key.Key("f7"),
	tcell.KeyF8:  key.Key("f8"),
	tcell.KeyF9:  key.Key("f9"),
	tcell.KeyF10: key.Key("f10"),
	tcell.KeyF11: key.Key("f11"),
	tcell.KeyF12: key.Key("f12"),

	tcell.KeyInsert: key.Key("insert"),
	tcell.KeyDelete: key.Key("delete"),
	tcell.KeyHome:   key.Key("home"),
	tcell.KeyEnd:    key.Key("end"),
	tcell.KeyPgUp:   key.Key("pgup"),
	tcell.KeyPgDn:   key.Key("pgdn"),

	tcell.KeyUp:    key.Key("up"),
	tcell.KeyDown:  key.Key("down"),
	tcell.KeyLeft:  key.Key("left"),
	tcell.KeyRight: key.Key("right"),

	tcell.KeyCtrlA:      key.Key("c-a"),
	tcell.KeyCtrlB:      key.Key("c-b"),
	tcell.KeyCtrlC:      key.Key("c-c"),
	tcell.KeyCtrlD:      key.Key("c-d"),
	tcell.KeyCtrlE:      key.Key("c-e"),
	tcell.KeyCtrlF:      key.Key("c-f"),
	tcell.KeyCtrlG:      key.Key("c-g"),
	tcell.KeyBackspace:  key.Key("backspace"),
	tcell.KeyTab:        key.Key("tab"),
	tcell.KeyBacktab:    key.Key("backtab"),
	tcell.KeyCtrlJ:      key.Key("c-j"),
	tcell.KeyCtrlK:      key.Key("c-k"),
	tcell.KeyCtrlL:      key.Key("c-l"),
	tcell.KeyEnter:      key.Key("enter"),
	tcell.KeyCtrlN:      key.Key("c-n"),
	tcell.KeyCtrlO:      key.Key("c-o"),
	tcell.KeyCtrlP:      key.Key("c-p"),
	tcell.KeyCtrlQ:      key.Key("c-q"),
	tcell.KeyCtrlR:      key.Key("c-r"),
	tcell.KeyCtrlS:      key.Key("c-s"),
	tcell.KeyCtrlT:      key.Key("c-t"),
	tcell.KeyCtrlU:      key.Key("c-u"),
	tcell.KeyCtrlV:      key.Key("c-v"),
	tcell.KeyCtrlW:      key.Key("c-w"),
	tcell.KeyCtrlX:      key.Key("c-x"),
	tcell.KeyCtrlY:      key.Key("c-y"),
	tcell.KeyCtrlZ:      key.Key("c-z"),
	tcell.KeyEsc:        key.Key("escape"),
	tcell.KeyBackspace2: key.Key("backspace2"),
}
