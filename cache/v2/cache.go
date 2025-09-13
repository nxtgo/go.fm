package cache

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/nxtgo/gce"
	"go.fm/lastfm/v2/types"
)

type Cache struct {
	User       *gce.Cache[string, types.UserGetInfo]
	Member     *gce.Cache[snowflake.ID, map[snowflake.ID]string]
	Cover      *gce.Cache[string, string]
	TopAlbums  *gce.Cache[string, types.UserGetTopAlbums]
	TopArtists *gce.Cache[string, types.UserGetTopArtists]
	TopTracks  *gce.Cache[string, types.UserGetTopTracks]
	Tracks     *gce.Cache[string, types.UserGetRecentTracks]
	Plays      *gce.Cache[string, int]
}

func NewCache() *Cache {
	return &Cache{
		User: gce.New[string, types.UserGetInfo](
			gce.WithDefaultTTL(time.Minute),
			gce.WithMaxEntries(50_000),
		),
		Member: gce.New[snowflake.ID, map[snowflake.ID]string](
			gce.WithDefaultTTL(time.Minute*10),
			gce.WithMaxEntries(2000),
		),
		Cover: gce.New[string, string](
			gce.WithDefaultTTL(time.Hour*12),
			gce.WithMaxEntries(50_000),
		),
		TopAlbums: gce.New[string, types.UserGetTopAlbums](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		TopArtists: gce.New[string, types.UserGetTopArtists](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		TopTracks: gce.New[string, types.UserGetTopTracks](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		Tracks: gce.New[string, types.UserGetRecentTracks](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(1000),
		),
		Plays: gce.New[string, int](
			gce.WithDefaultTTL(time.Minute*15),
			gce.WithMaxEntries(50_000),
		),
	}
}

func (c *Cache) Close() {
	if c.User != nil {
		c.User.Close()
	}
	if c.Member != nil {
		c.Member.Close()
	}
	if c.Cover != nil {
		c.Cover.Close()
	}
	if c.TopAlbums != nil {
		c.TopAlbums.Close()
	}
	if c.TopArtists != nil {
		c.TopArtists.Close()
	}
	if c.TopTracks != nil {
		c.TopTracks.Close()
	}
	if c.Tracks != nil {
		c.Tracks.Close()
	}
	if c.Plays != nil {
		c.Plays.Close()
	}
}
