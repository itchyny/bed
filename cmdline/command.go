package cmdline

import "github.com/itchyny/bed/event"

type command struct {
	name      string
	fullname  string
	eventType event.Type
	rangeType rangeType
}

type rangeType int

const (
	rangeEmpty rangeType = 1 << iota
	rangeCount
	rangeBoth
)

func (rt rangeType) allows(r *event.Range) bool {
	switch {
	case r == nil:
		return rt&rangeEmpty != 0
	case r.To == nil:
		return rt&rangeCount != 0
	default:
		return rt&rangeBoth != 0
	}
}

var commands = []command{
	{"e[dit]", "edit", event.Edit, rangeEmpty},
	{"ene[w]", "enew", event.Enew, rangeEmpty},
	{"new", "new", event.New, rangeEmpty},
	{"vne[w]", "vnew", event.Vnew, rangeEmpty},
	{"on[ly]", "only", event.Only, rangeEmpty},
	{"winc[md]", "wincmd", event.Wincmd, rangeEmpty},

	{"go[to]", "goto", event.CursorGoto, rangeCount},
	{"%", "%", event.CursorGoto, rangeCount},

	{"u[ndo]", "undo", event.Undo, rangeEmpty},
	{"red[o]", "redo", event.Redo, rangeEmpty},

	{"pw[d]", "pwd", event.Pwd, rangeEmpty},
	{"cd", "cd", event.Chdir, rangeEmpty},
	{"chd[ir]", "chdir", event.Chdir, rangeEmpty},
	{"exi[t]", "exit", event.Quit, rangeEmpty},
	{"q[uit]", "quit", event.Quit, rangeEmpty},
	{"qa[ll]", "qall", event.QuitAll, rangeEmpty},
	{"quita[ll]", "quitall", event.QuitAll, rangeEmpty},
	{"cq[uit]", "cquit", event.QuitErr, rangeEmpty},
	{"w[rite]", "write", event.Write, rangeEmpty | rangeBoth},
	{"wq", "wq", event.WriteQuit, rangeEmpty | rangeBoth},
	{"x[it]", "xit", event.WriteQuit, rangeEmpty | rangeBoth},
	{"xa[ll]", "xall", event.WriteQuit, rangeEmpty | rangeBoth},
}
