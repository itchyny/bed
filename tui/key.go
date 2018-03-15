package tui

import (
	"github.com/gdamore/tcell"

	. "github.com/itchyny/bed/common"
)

func eventToKey(event *tcell.EventKey) Key {
	if key, ok := keyMap[event.Key()]; ok {
		return key
	}
	return Key(event.Rune())
}

var keyMap = map[tcell.Key]Key{
	tcell.KeyF1:  Key("f1"),
	tcell.KeyF2:  Key("f2"),
	tcell.KeyF3:  Key("f3"),
	tcell.KeyF4:  Key("f4"),
	tcell.KeyF5:  Key("f5"),
	tcell.KeyF6:  Key("f6"),
	tcell.KeyF7:  Key("f7"),
	tcell.KeyF8:  Key("f8"),
	tcell.KeyF9:  Key("f9"),
	tcell.KeyF10: Key("f10"),
	tcell.KeyF11: Key("f11"),
	tcell.KeyF12: Key("f12"),

	tcell.KeyInsert: Key("insert"),
	tcell.KeyDelete: Key("delete"),
	tcell.KeyHome:   Key("home"),
	tcell.KeyEnd:    Key("end"),
	tcell.KeyPgUp:   Key("pgup"),
	tcell.KeyPgDn:   Key("pgdn"),

	tcell.KeyUp:    Key("up"),
	tcell.KeyDown:  Key("down"),
	tcell.KeyLeft:  Key("left"),
	tcell.KeyRight: Key("right"),

	tcell.KeyCtrlA:      Key("c-a"),
	tcell.KeyCtrlB:      Key("c-b"),
	tcell.KeyCtrlC:      Key("c-c"),
	tcell.KeyCtrlD:      Key("c-d"),
	tcell.KeyCtrlE:      Key("c-e"),
	tcell.KeyCtrlF:      Key("c-f"),
	tcell.KeyCtrlG:      Key("c-g"),
	tcell.KeyBackspace:  Key("backspace"),
	tcell.KeyTab:        Key("tab"),
	tcell.KeyCtrlJ:      Key("c-j"),
	tcell.KeyCtrlK:      Key("c-k"),
	tcell.KeyCtrlL:      Key("c-l"),
	tcell.KeyEnter:      Key("enter"),
	tcell.KeyCtrlN:      Key("c-n"),
	tcell.KeyCtrlO:      Key("c-o"),
	tcell.KeyCtrlP:      Key("c-p"),
	tcell.KeyCtrlQ:      Key("c-q"),
	tcell.KeyCtrlR:      Key("c-r"),
	tcell.KeyCtrlS:      Key("c-s"),
	tcell.KeyCtrlT:      Key("c-t"),
	tcell.KeyCtrlU:      Key("c-u"),
	tcell.KeyCtrlV:      Key("c-v"),
	tcell.KeyCtrlW:      Key("c-w"),
	tcell.KeyCtrlX:      Key("c-x"),
	tcell.KeyCtrlY:      Key("c-y"),
	tcell.KeyCtrlZ:      Key("c-z"),
	tcell.KeyEsc:        Key("escape"),
	tcell.KeyBackspace2: Key("backspace2"),
}
