package botinfo

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go.fm/constants"
	"go.fm/types/cmd"
)

type Command struct{}

func (Command) Data() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "botinfo",
		Description: "display go.fm's info",
		IntegrationTypes: []discord.ApplicationIntegrationType{
			discord.ApplicationIntegrationTypeGuildInstall,
			discord.ApplicationIntegrationTypeUserInstall,
		},
	}
}

func (Command) Handle(e *events.ApplicationCommandInteractionCreate, ctx cmd.CommandContext) {
	reply := ctx.Reply(e)

	if err := reply.Defer(); err != nil {
		_ = ctx.Error(e, constants.ErrorAcknowledgeCommand)
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	lastFMUsers, err := ctx.Database.GetUserCount(ctx.Context)
	if err != nil {
		lastFMUsers = 0
	}

	branch, commit, message := getGitInfo()

	stats := fmt.Sprintf(
		"```\n"+
			"registered last.fm users: %d\n"+
			"goroutines: %d\n"+
			"memory Usage:\n"+
			"  - alloc: %.2f MB\n"+
			"  - total: %.2f MB\n"+
			"  - sys: %.2f MB\n"+
			"uptime: %s\n"+
			"git:\n"+
			"  - branch: %s\n"+
			"  - commit: %s\n"+
			"  - message: %s\n"+
			"```",
		lastFMUsers,
		runtime.NumGoroutine(),
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		time.Since(ctx.Start).Truncate(time.Second),
		branch,
		commit,
		message,
	)

	reply.Content(stats).Edit()
}

func getGitInfo() (branch, commit, message string) {
	branch = runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	commit = runGitCommand("rev-parse", "--short", "HEAD")
	message = runGitCommand("log", "-1", "--pretty=%B")
	return
}

func runGitCommand(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
