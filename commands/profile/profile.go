package profile

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"go.fm/constants"
	lfm "go.fm/lastfm/v2"
	"go.fm/lastfm/v2/types"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "profile",
		Description: "display a last.fm user info",
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

	username, err := ctx.GetUser(e)
	if err != nil {
		_ = ctx.Error(e, constants.ErrorNotRegistered)
		return
	}

	user, err := ctx.LastFM.User.GetInfo(lfm.P{"user": username})
	if err != nil {
		_ = ctx.Error(e, constants.ErrorUserNotFound)
		return
	}

	realName := user.RealName
	if realName == "" {
		realName = user.Name
	}

	var favTrack *types.UserGetTopTracks
	var trackName, trackURL string

	topTracks, err := ctx.LastFM.User.GetTopTracks(lfm.P{"user": username, "limit": 1})
	if err == nil && len(topTracks.Tracks) > 0 {
		favTrack = topTracks
		trackName = favTrack.Tracks[0].Name
		trackURL = favTrack.Tracks[0].Url
	} else {
		trackName = "none"
		trackURL = ""
	}

	var favArtist *types.UserGetTopArtists
	var artistName, artistURL string

	topArtists, err := ctx.LastFM.User.GetTopArtists(lfm.P{"user": username, "limit": 1})
	if err == nil && len(topArtists.Artists) > 0 {
		favArtist = topArtists
		artistName = favArtist.Artists[0].Name
		artistURL = favArtist.Artists[0].Url
	} else {
		artistName = "none"
		artistURL = ""
	}

	var favAlbum *types.UserGetTopAlbums
	var albumName, albumURL string

	topAlbums, err := ctx.LastFM.User.GetTopAlbums(lfm.P{"user": username, "limit": 1})
	if err == nil && len(topAlbums.Albums) > 0 {
		favAlbum = topAlbums
		albumName = favAlbum.Albums[0].Name
		albumURL = favAlbum.Albums[0].Url
	} else {
		albumName = "none"
		albumURL = ""
	}

	avatar := ""
	if len(user.Images) > 0 {
		avatar = user.Images[len(user.Images)-1].Url
	}
	if avatar == "" {
		avatar = "https://lastfm.freetls.fastly.net/i/u/avatar170s/818148bf682d429dc215c1705eb27b98.png"
	}
	if dot := strings.LastIndex(avatar, "."); dot != -1 {
		avatar = avatar[:dot] + ".gif"
	}

	component := discord.NewContainer(
		discord.NewSection(
			discord.NewTextDisplayf("## [%s](%s)", realName, user.Url),
			discord.NewTextDisplayf("-# *__@%s__*\nsince <t:%s:D>", user.Name, user.Registered.Unixtime),
			discord.NewTextDisplayf("**%s** total scrobbles", user.PlayCount),
		).WithAccessory(discord.NewThumbnail(avatar)),
		discord.NewSmallSeparator(),
		discord.NewTextDisplay(
			fmt.Sprintf("-# *Favorite album* \\💿\n[**%s**](%s)\n", albumName, albumURL)+
				fmt.Sprintf("-# *Favorite artist* \\🎤\n[**%s**](%s)\n", artistName, artistURL)+
				fmt.Sprintf("-# *Favorite track* \\🎵\n[**%s**](%s)\n", trackName, trackURL),
		),
		discord.NewSmallSeparator(),
		discord.NewTextDisplayf(
			"\\🎤 **%s** artists\n\\💿 **%s** albums\n\\🎵 **%s** unique tracks",
			user.ArtistCount,
			user.AlbumCount,
			user.TrackCount,
		),
	).WithAccentColor(0x00ADD8)

	reply.Flags(discord.MessageFlagIsComponentsV2).Component(component).Edit()
}
