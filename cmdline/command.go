package cmdline

import . "github.com/itchyny/bed/core"

type command struct {
	name      string
	eventType EventType
}

var commands = []command{
	{"exi[t]", EventQuit},
	{"q[uit]", EventQuit},
	{"qa[ll]", EventQuit},
	{"w[rite]", EventWrite},
	{"wq", EventWriteQuit},
	{"x[it]", EventWriteQuit},
	{"xa[ll]", EventWriteQuit},
}
