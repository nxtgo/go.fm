package setuser

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"go.fm/commands"
	"go.fm/db"
	"go.fm/pkg/reply"
)

var data = api.CreateCommandData{
	Name:        "set-user",
	Description: "set your last.fm username for go.fm",
	Options: discord.CommandOptions{
		discord.NewStringOption("username", "your last.fm username", true),
	},
}

var options struct {
	Username string `discord:"username"`
}

func handler(c *commands.CommandContext) error {
	return c.Reply.AutoDefer(func(edit *reply.EditBuilder) error {
		if err := c.Data.Options.Unmarshal(&options); err != nil {
			return err
		}

		userID := c.Data.Event.Member.User.ID

		user, err := c.Query.GetUserByLastFM(c.Context, options.Username)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check username: %w", err)
		}
		if err == nil && user.UserID != userID.String() {
			return fmt.Errorf("the username **%s** is already linked to another discord user", options.Username)
		}
		if err == nil && user.UserID == userID.String() {
			return fmt.Errorf("your username is already set to **%s**", options.Username)
		}

		err = c.Query.UpsertUser(c.Context, db.UpsertUserParams{
			UserID:         userID.String(),
			LastfmUsername: options.Username,
		})
		if err != nil {
			return fmt.Errorf("failed to update username: %w", err)
		}

		_, err = edit.Embed(reply.SuccessEmbed(fmt.Sprintf("updated your username to **%s**", options.Username))).Send()
		return err
	})
}

func init() {
	commands.Register(data, handler)
}
