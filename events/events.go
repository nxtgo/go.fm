package events

import (
	"github.com/diamondburned/arikawa/v3/gateway"
	"go.fm/zlog"
)

var Events []any

func init() {
	Events = append(Events, EventReady)
}

func EventReady(c *gateway.ReadyEvent) {
	zlog.Log.WithFields(
		zlog.F{
			"tag":    c.User.Tag(),
			"guilds": len(c.Guilds),
		},
	).Info("client is ready")
}
