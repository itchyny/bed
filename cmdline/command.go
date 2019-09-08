package cmdline

import "github.com/itchyny/bed/event"

type command struct {
	name      string
	eventType event.Type
}

var commands = []command{
	{"e[dit]", event.Edit},
	{"ene[w]", event.Enew},
	{"new", event.New},
	{"vne[w]", event.Vnew},
	{"winc[md]", event.Wincmd},

	{"u[ndo]", event.Undo},
	{"red[o]", event.Redo},

	{"exi[t]", event.Quit},
	{"q[uit]", event.Quit},
	{"qa[ll]", event.QuitAll},
	{"quita[ll]", event.QuitAll},
	{"w[rite]", event.Write},
	{"wq", event.WriteQuit},
	{"x[it]", event.WriteQuit},
	{"xa[ll]", event.WriteQuit},
}
