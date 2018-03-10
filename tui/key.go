package tui

import (
	"github.com/nsf/termbox-go"

	. "github.com/itchyny/bed/core"
)

func eventToKey(event termbox.Event) Key {
	if key, ok := keyMap[event.Key]; ok {
		return key
	}
	return Key(event.Ch)
}

var keyMap = map[termbox.Key]Key{
	termbox.KeyF1:  Key("f1"),
	termbox.KeyF2:  Key("f2"),
	termbox.KeyF3:  Key("f3"),
	termbox.KeyF4:  Key("f4"),
	termbox.KeyF5:  Key("f5"),
	termbox.KeyF6:  Key("f6"),
	termbox.KeyF7:  Key("f7"),
	termbox.KeyF8:  Key("f8"),
	termbox.KeyF9:  Key("f9"),
	termbox.KeyF10: Key("f10"),
	termbox.KeyF11: Key("f11"),
	termbox.KeyF12: Key("f12"),

	termbox.KeyInsert: Key("insert"),
	termbox.KeyDelete: Key("delete"),
	termbox.KeyHome:   Key("home"),
	termbox.KeyEnd:    Key("end"),
	termbox.KeyPgup:   Key("pgup"),
	termbox.KeyPgdn:   Key("pgdn"),

	termbox.KeyArrowUp:    Key("up"),
	termbox.KeyArrowDown:  Key("down"),
	termbox.KeyArrowLeft:  Key("left"),
	termbox.KeyArrowRight: Key("right"),

	termbox.KeyCtrlA:          Key("c-a"),
	termbox.KeyCtrlB:          Key("c-b"),
	termbox.KeyCtrlC:          Key("c-c"),
	termbox.KeyCtrlD:          Key("c-d"),
	termbox.KeyCtrlE:          Key("c-e"),
	termbox.KeyCtrlF:          Key("c-f"),
	termbox.KeyCtrlG:          Key("c-g"),
	termbox.KeyBackspace:      Key("backspace"),
	termbox.KeyTab:            Key("tab"),
	termbox.KeyCtrlJ:          Key("c-j"),
	termbox.KeyCtrlK:          Key("c-k"),
	termbox.KeyCtrlL:          Key("c-l"),
	termbox.KeyEnter:          Key("enter"),
	termbox.KeyCtrlN:          Key("c-n"),
	termbox.KeyCtrlO:          Key("c-o"),
	termbox.KeyCtrlP:          Key("c-p"),
	termbox.KeyCtrlQ:          Key("c-q"),
	termbox.KeyCtrlR:          Key("c-r"),
	termbox.KeyCtrlS:          Key("c-s"),
	termbox.KeyCtrlT:          Key("c-t"),
	termbox.KeyCtrlU:          Key("c-u"),
	termbox.KeyCtrlV:          Key("c-v"),
	termbox.KeyCtrlW:          Key("c-w"),
	termbox.KeyCtrlX:          Key("c-x"),
	termbox.KeyCtrlY:          Key("c-y"),
	termbox.KeyCtrlZ:          Key("c-z"),
	termbox.KeyEsc:            Key("escape"),
	termbox.KeyCtrl4:          Key("c-4"),
	termbox.KeyCtrlRsqBracket: Key("c-]"),
	termbox.KeyCtrl6:          Key("c-6"),
	termbox.KeyCtrlSlash:      Key("c-slash"),
	termbox.KeySpace:          Key("space"),
	termbox.KeyBackspace2:     Key("backspace2"),
}
