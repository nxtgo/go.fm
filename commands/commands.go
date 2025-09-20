package commands

import (
	"context"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/state"
	"go.fm/db"
	"go.fm/pkg/reply"
	"go.fm/zlog"
)

var allCommands = []api.CreateCommandData{}
var registry = map[string]CommandHandler{}

func Register(meta api.CreateCommandData, handler CommandHandler) {
	zlog.Log.Debugf("registered command %s", meta.Name)

	allCommands = append(allCommands, meta)
	registry[meta.Name] = handler
}

func RegisterCommands(r *cmdroute.Router, st *state.State, q *db.Queries) {
	for name, handler := range registry {
		h := handler
		r.AddFunc(name, func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
			replyManager := reply.New(st, data.Event)
			err := h(&CommandContext{
				Context: ctx,
				Data:    data,
				State:   st,
				Reply:   replyManager,
				Query:   q,
			})

			if err != nil {
				replyManager.QuickEmbed(reply.ErrorEmbed(err.Error()))
			}

			return nil
		})
	}
}

func Sync(st *state.State) error {
	zlog.Log.Debug("synced commands")

	return cmdroute.OverwriteCommands(st, allCommands)
}
