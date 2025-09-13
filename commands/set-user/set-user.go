package setuser

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/nxtgo/zlog"

	"go.fm/constants"
	"go.fm/db"
	lfm "go.fm/lastfm/v2"
	"go.fm/logger"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "set-user",
		Description: "link your last.fm username to your Discord account",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "username",
				Description: "your last.fm username",
				Required:    true,
			},
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	username := e.SlashCommandInteractionData().String("username")
	discordID := e.User().ID.String()

	_, err := ctx.LastFM.User.GetInfo(lfm.P{"user": username})
	if err != nil {
		_ = ctx.Error(e, constants.ErrorUserNotFound)
		return
	}

	existing, err := ctx.Database.GetUserByUsername(ctx.Context, username)
	if err == nil {
		if existing.DiscordID != discordID {
			_ = ctx.Error(e, constants.ErrorAlreadyLinked)
			return
		}
		if existing.LastfmUsername == username {
			_ = ctx.Error(e, fmt.Sprintf(constants.ErrorUsernameAlreadySet, username))
			return
		}
	}

	if errors.Is(err, sql.ErrNoRows) || existing.DiscordID == discordID {
		if dbErr := ctx.Database.UpsertUser(ctx.Context, db.UpsertUserParams{
			DiscordID:      discordID,
			LastfmUsername: username,
		}); dbErr != nil {
			logger.Log.Errorw("failed to upsert user", zlog.F{"gid": e.GuildID().String(), "uid": discordID}, dbErr)
			_ = ctx.Error(e, constants.ErrorSetUsername)
			return
		}

		reply.Content("your last.fm username has been set to **%s**", username).Edit()
	}
}
