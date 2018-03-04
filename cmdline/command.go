package cmdline

import "github.com/itchyny/bed/core"

type command struct {
	name      string
	eventType core.EventType
}

var commands = []command{
	{"exi[t]", core.EventQuit},
	{"q[uit]", core.EventQuit},
	{"qa[ll]", core.EventQuit},
	{"w[rite]", core.EventWrite},
	{"wq", core.EventWriteQuit},
	{"x[it]", core.EventWriteQuit},
	{"xa[ll]", core.EventWriteQuit},
}
