package cache

import (
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/nxtgo/gce"
	"go.fm/types/lastfm"
)

// i think i gotta rewrite ts again :wilted_rose:
// cache isn't last.fm enough.

type LastFMCache struct {
	user       *gce.Cache[string, *lastfm.UserInfoResponse]
	track      *gce.Cache[string, *lastfm.TrackInfoResponse]
	topArtists *gce.Cache[string, *lastfm.TopArtistsResponse]
	topAlbums  *gce.Cache[string, *lastfm.TopAlbumsResponse]
	topTracks  *gce.Cache[string, *lastfm.TopTracksResponse]
	plays      *gce.Cache[string, int]
	member     *gce.Cache[snowflake.ID, map[snowflake.ID]string]
}

func NewLastFMCache() *LastFMCache {
	commonCleanup := gce.WithCleanupInterval(30 * time.Second)
	commonShards := gce.WithShardCount(64)

	return &LastFMCache{
		user: gce.New[string, *lastfm.UserInfoResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(5*time.Minute),
			gce.WithMaxEntries(10_000),
		),
		track: gce.New[string, *lastfm.TrackInfoResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(30*time.Minute),
			gce.WithMaxEntries(50_000),
		),
		topArtists: gce.New[string, *lastfm.TopArtistsResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(15*time.Minute),
			gce.WithMaxEntries(10_000),
		),
		topAlbums: gce.New[string, *lastfm.TopAlbumsResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(15*time.Minute),
			gce.WithMaxEntries(10_000),
		),
		topTracks: gce.New[string, *lastfm.TopTracksResponse](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(15*time.Minute),
			gce.WithMaxEntries(10_000),
		),
		plays: gce.New[string, int](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(10*time.Minute),
			gce.WithMaxEntries(100_000),
		),
		member: gce.New[snowflake.ID, map[snowflake.ID]string](
			commonShards,
			commonCleanup,
			gce.WithDefaultTTL(5*time.Minute),
			gce.WithMaxEntries(1000),
		),
	}
}

func (c *LastFMCache) Stats() string {
	stats := "```\n"

	caches := map[string]any{
		"user cache":   c.user,
		"track cache":  c.track,
		"plays cache":  c.plays,
		"member cache": c.member,
		"top artists":  c.topArtists,
		"top albums":   c.topAlbums,
		"top tracks":   c.topTracks,
	}

	for name, cacheObj := range caches {
		if cacheObj == nil {
			continue
		}

		// lol.
		var s gce.Stats
		switch cache := cacheObj.(type) {
		case *gce.Cache[string, any]:
			s = cache.Stats()
		case *gce.Cache[string, int]:
			s = cache.Stats()
		case *gce.Cache[snowflake.ID, map[snowflake.ID]string]:
			s = cache.Stats()
		case *gce.Cache[string, *lastfm.UserInfoResponse]:
			s = cache.Stats()
		case *gce.Cache[string, *lastfm.TrackInfoResponse]:
			s = cache.Stats()
		case *gce.Cache[string, *lastfm.TopArtistsResponse]:
			s = cache.Stats()
		case *gce.Cache[string, *lastfm.TopAlbumsResponse]:
			s = cache.Stats()
		case *gce.Cache[string, *lastfm.TopTracksResponse]:
			s = cache.Stats()
		default:
			continue
		}

		stats += fmt.Sprintf(
			"%-15s | hits: %6d | misses: %6d | loads: %6d | evictions: %6d | size: %6d\n",
			name, s.Hits, s.Misses, s.Loads, s.Evictions, s.CurrentSize,
		)
	}

	stats += "```"
	return stats
}

// Close all caches
func (c *LastFMCache) Close() {
	if c.user != nil {
		c.user.Close()
	}
	if c.track != nil {
		c.track.Close()
	}
	if c.topArtists != nil {
		c.topArtists.Close()
	}
	if c.topAlbums != nil {
		c.topAlbums.Close()
	}
	if c.topTracks != nil {
		c.topTracks.Close()
	}
	if c.plays != nil {
		c.plays.Close()
	}
	if c.member != nil {
		c.member.Close()
	}
}
