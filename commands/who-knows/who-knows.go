package whoknows

import (
	"fmt"
	"sort"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	lfm "go.fm/lastfm/v2"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "who-knows",
		Description: "see who in this guild has listened to a track/artist/album the most",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "type",
				Description: "artist, track or album",
				Choices: []discord.ApplicationCommandOptionChoiceString{
					{Name: "artist", Value: "artist"},
					{Name: "track", Value: "track"},
					{Name: "album", Value: "album"},
				},
				Required: true,
			},
			discord.ApplicationCommandOptionString{
				Name:        "name",
				Description: "name of the artist/track/album",
				Required:    false,
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

	tType := e.SlashCommandInteractionData().String("type")
	name, defined := e.SlashCommandInteractionData().OptString("name")

	var img string
	if !defined {
		currentUser, err := ctx.Database.GetUser(ctx.Context, e.Member().User.ID.String())
		if err != nil {
			_ = ctx.Error(e, constants.ErrorGetUser)
			return
		}

		tracks, err := ctx.LastFM.User.GetRecentTracks(lfm.P{"user": currentUser, "limit": 1})
		if err != nil || len(tracks.Tracks) == 0 || tracks.Tracks[0].NowPlaying != "true" {
			_ = ctx.Error(e, constants.ErrorFetchCurrentTrack)
			return
		}

		current := tracks.Tracks[0]
		img = current.Images[len(current.Images)-1].Url

		switch tType {
		case "artist":
			name = current.Artist.Name
		case "track":
			name = current.Name
		case "album":
			name = current.Album.Name
		}
	}

	users, err := ctx.LastFM.User.GetUsersByGuild(ctx.Context, e, ctx.Database)
	if err != nil {
		_ = ctx.Error(e, constants.ErrorUnexpected)
		return
	}

	type result struct {
		UserID    string
		Username  string
		PlayCount int
	}

	var (
		results []result
		mu      sync.Mutex
		wg      sync.WaitGroup
		sem     = make(chan struct{}, 10)
	)

	for id, username := range users {
		idCopy, usernameCopy := id.String(), username
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			count, err := ctx.LastFM.User.GetPlays(lfm.P{
				"user":  usernameCopy,
				"name":  name,
				"type":  tType,
				"limit": 1000,
			})
			if err != nil || count == 0 {
				return
			}

			mu.Lock()
			results = append(results, result{
				UserID:    idCopy,
				Username:  usernameCopy,
				PlayCount: count,
			})
			mu.Unlock()
		})
	}

	wg.Wait()
	close(sem)

	if len(results) == 0 {
		_ = ctx.Error(e, constants.ErrorNoListeners)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].PlayCount > results[j].PlayCount
	})

	list := ""
	for i, r := range results {
		if i >= 10 {
			break
		}
		list += fmt.Sprintf("%d. [%s](<https://www.last.fm/user/%s>) (<@%s>) — %d plays\n",
			i+1, r.Username, r.Username, r.UserID, r.PlayCount)
	}

	embed := ctx.QuickEmbed(name, list)
	embed.Author = &discord.EmbedAuthor{Name: "who listened more to..."}
	if img != "" {
		embed.Thumbnail = &discord.EmbedResource{URL: img}
	}

	_ = reply.Embed(embed).Edit()
}
