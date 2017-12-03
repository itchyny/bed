package tui

import (
	"github.com/itchyny/bed/core"
	"github.com/nsf/termbox-go"
)

func eventToKey(event termbox.Event) core.Key {
	if key, ok := keyMap[event.Key]; ok {
		return key
	}
	return core.Key(event.Ch)
}

var keyMap = map[termbox.Key]core.Key{
	termbox.KeyF1:  core.Key("f1"),
	termbox.KeyF2:  core.Key("f2"),
	termbox.KeyF3:  core.Key("f3"),
	termbox.KeyF4:  core.Key("f4"),
	termbox.KeyF5:  core.Key("f5"),
	termbox.KeyF6:  core.Key("f6"),
	termbox.KeyF7:  core.Key("f7"),
	termbox.KeyF8:  core.Key("f8"),
	termbox.KeyF9:  core.Key("f9"),
	termbox.KeyF10: core.Key("f10"),
	termbox.KeyF11: core.Key("f11"),
	termbox.KeyF12: core.Key("f12"),

	termbox.KeyInsert: core.Key("insert"),
	termbox.KeyDelete: core.Key("delete"),
	termbox.KeyHome:   core.Key("home"),
	termbox.KeyEnd:    core.Key("end"),
	termbox.KeyPgup:   core.Key("pgup"),
	termbox.KeyPgdn:   core.Key("pgdn"),

	termbox.KeyArrowUp:    core.Key("up"),
	termbox.KeyArrowDown:  core.Key("down"),
	termbox.KeyArrowLeft:  core.Key("left"),
	termbox.KeyArrowRight: core.Key("right"),

	termbox.KeyCtrlA:          core.Key("c-a"),
	termbox.KeyCtrlB:          core.Key("c-b"),
	termbox.KeyCtrlC:          core.Key("c-c"),
	termbox.KeyCtrlD:          core.Key("c-d"),
	termbox.KeyCtrlE:          core.Key("c-e"),
	termbox.KeyCtrlF:          core.Key("c-f"),
	termbox.KeyCtrlG:          core.Key("c-g"),
	termbox.KeyBackspace:      core.Key("backspace"),
	termbox.KeyTab:            core.Key("tab"),
	termbox.KeyCtrlJ:          core.Key("c-j"),
	termbox.KeyCtrlK:          core.Key("c-k"),
	termbox.KeyCtrlL:          core.Key("c-l"),
	termbox.KeyEnter:          core.Key("enter"),
	termbox.KeyCtrlN:          core.Key("c-n"),
	termbox.KeyCtrlO:          core.Key("c-o"),
	termbox.KeyCtrlP:          core.Key("c-p"),
	termbox.KeyCtrlQ:          core.Key("c-q"),
	termbox.KeyCtrlR:          core.Key("c-r"),
	termbox.KeyCtrlS:          core.Key("c-s"),
	termbox.KeyCtrlT:          core.Key("c-t"),
	termbox.KeyCtrlU:          core.Key("c-u"),
	termbox.KeyCtrlV:          core.Key("c-v"),
	termbox.KeyCtrlW:          core.Key("c-w"),
	termbox.KeyCtrlX:          core.Key("c-x"),
	termbox.KeyCtrlY:          core.Key("c-y"),
	termbox.KeyCtrlZ:          core.Key("c-z"),
	termbox.KeyEsc:            core.Key("escape"),
	termbox.KeyCtrl4:          core.Key("c-4"),
	termbox.KeyCtrlRsqBracket: core.Key("c-]"),
	termbox.KeyCtrl6:          core.Key("c-6"),
	termbox.KeyCtrlSlash:      core.Key("c-slash"),
	termbox.KeySpace:          core.Key("space"),
	termbox.KeyBackspace2:     core.Key("backspace2"),
}
