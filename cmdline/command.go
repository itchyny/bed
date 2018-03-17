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

	{"exi[t]", EventQuit},
	{"q[uit]", EventQuit},
	{"qa[ll]", EventQuit},
	{"w[rite]", EventWrite},
	{"wq", EventWriteQuit},
	{"x[it]", EventWriteQuit},
	{"xa[ll]", EventWriteQuit},
}
