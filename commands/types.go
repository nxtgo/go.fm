package commands

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/state"
	"go.fm/db"
	"go.fm/pkg/reply"
)

type CommandContext struct {
	Context context.Context
	Data    cmdroute.CommandData
	State   *state.State
	Reply   *reply.ResponseManager
	Query   *db.Queries
}

type CommandHandler func(c *CommandContext) error

func (ctx *CommandContext) GetUserOrFallback() {
	panic("GetUserOrFallback shouldn't be called, not implemented yet, lol")
}
