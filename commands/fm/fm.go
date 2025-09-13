package fm

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	lfm "go.fm/lastfm/v2"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "fm",
		Description: "get an user's current track",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
		Options: []discord.ApplicationCommandOption{
			cmd.UserOption,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	user, err := ctx.GetUser(e)
	if err != nil {
		_ = ctx.Error(e, err.Error())
		return
	}

	data, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": user, "limit": 1})
	if err != nil {
		_ = ctx.Error(e, constants.ErrorFetchCurrentTrack)
		return
	}

	if len(data.Tracks) == 0 {
		_ = ctx.Error(e, constants.ErrorNoTracks)
		return
	}

	track := data.Tracks[0]
	if track.NowPlaying != "true" {
		_ = ctx.Error(e, constants.ErrorNotPlaying)
		return
	}

	embed := ctx.QuickEmbed(
		track.Name,
		fmt.Sprintf("by **%s**\n-# *at %s*", track.Artist.Name, track.Album.Name),
	)
	embed.Author = &discord.EmbedAuthor{
		Name: fmt.Sprintf("%s's current track", user),
		URL:  fmt.Sprintf("https://www.last.fm/user/%s", user),
	}
	embed.URL = track.Url
	if len(track.Images) > 0 {
		embed.Thumbnail = &discord.EmbedResource{
			URL: track.Images[len(track.Images)-1].Url,
		}
	}

	_ = reply.Embed(embed).Edit()
}
