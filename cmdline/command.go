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
	{"on[ly]", event.Only},
	{"winc[md]", event.Wincmd},

	{"go[to]", event.CursorGoto},
	{"%", event.CursorGoto},

	{"u[ndo]", event.Undo},
	{"red[o]", event.Redo},

	{"pw[d]", event.Pwd},
	{"cd", event.Chdir},
	{"chd[ir]", event.Chdir},
	{"exi[t]", event.Quit},
	{"q[uit]", event.Quit},
	{"qa[ll]", event.QuitAll},
	{"quita[ll]", event.QuitAll},
	{"cq[uit]", event.QuitErr},
	{"w[rite]", event.Write},
	{"wq", event.WriteQuit},
	{"x[it]", event.WriteQuit},
	{"xa[ll]", event.WriteQuit},
}
