package cmdline

import . "github.com/itchyny/bed/common"

type command struct {
	name      string
	eventType EventType
}

var commands = []command{
	{"e[dit]", EventEdit},
	{"new", EventNew},
	{"vne[w]", EventVnew},
	{"winc[md]", EventWincmd},

	{"exi[t]", EventQuit},
	{"q[uit]", EventQuit},
	{"qa[ll]", EventQuitAll},
	{"quita[ll]", EventQuitAll},
	{"w[rite]", EventWrite},
	{"wq", EventWriteQuit},
	{"x[it]", EventWriteQuit},
	{"xa[ll]", EventWriteQuit},
}
