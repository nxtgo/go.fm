package stats

import (
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"go.fm/commands"
	"runtime"
	"time"
)

var startTime = time.Now()

var data = api.CreateCommandData{
	Name:        "stats",
	Description: "display bot's stats",
}

func handler(c *commands.CommandContext) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime).Round(time.Second)

	embed := discord.Embed{Description: fmt.Sprintf(
		"**uptime:** %s\n"+
			"**goroutines:** %d\n"+
			"**memory:** %.2f mb\n"+
			"**heap:** %.2f mb\n"+
			"**gc runs:** %d\n"+
			"**go version:** %s\n"+
			"**platform:** %s/%s",
		uptime,
		runtime.NumGoroutine(),
		float64(m.Alloc)/(1024*1024),
		float64(m.HeapAlloc)/(1024*1024),
		m.NumGC,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)}

	return c.Reply.QuickEmbed(embed)
}

func init() {
	commands.Register(data, handler)
}
