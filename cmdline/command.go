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
	{"x[it]", core.EventQuit},
	{"xa[ll]", core.EventQuit},
}
