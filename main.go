package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo/bot"
	dgocache "github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/snowflake/v2"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/gateway"
	"github.com/nxtgo/env"

	"go.fm/cache/v2"
	"go.fm/commands"
	"go.fm/db"
	"go.fm/lastfm/v2"
	"go.fm/logger"
	"go.fm/types/cmd"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed db/sql/schema.sql
var ddl string

var (
	globalCmds bool
	dbPath     string
)

func init() {
	if err := env.Load(""); err != nil {
		logger.Log.Fatalf("failed loading environment: %v", err)
	}

	flag.BoolVar(&globalCmds, "g", true, "upload global commands to discord")
	flag.StringVar(&dbPath, "db", "database.db", "path to the SQLite database file")
	flag.Parse()
}

func main() {
	ctx := context.Background()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Log.Fatal("missing DISCORD_TOKEN environment variable")
	}

	lfmCache := cache.NewCache()
	lfm := lfm.New(os.Getenv("LASTFM_API_KEY"), lfmCache)
	defer lfmCache.Close()

	closeConnection, database := initDatabase(ctx, dbPath)
	defer closeConnection()

	cmdCtx := cmd.CommandContext{
		LastFM:   lfm,
		Database: database,
		Cache:    lfmCache,
		Context:  ctx,
	}
	commands.InitDependencies(cmdCtx)

	client := initDiscordClient(token)
	defer client.Close(context.TODO())

	if err := client.OpenGateway(context.TODO()); err != nil {
		logger.Log.Fatalf("failed to open gateway: %v", err)
	}

	if globalCmds {
		uploadGlobalCommands(*client)
	} else {
		uploadGuildCommands(*client)
	}

	waitForShutdown()
}

func initDatabase(ctx context.Context, path string) (func() error, *db.Queries) {
	dbConn, err := sql.Open("sqlite3", path)
	if err != nil {
		logger.Log.Fatalf("failed opening database: %v", err)
	}

	if _, err := dbConn.ExecContext(ctx, ddl); err != nil {
		logger.Log.Fatalf("failed executing schema: %v", err)
	}

	_, err = db.Prepare(ctx, dbConn)
	if err != nil {
		logger.Log.Fatalf("failed preparing queries: %v", err)
	}

	return dbConn.Close, db.New(dbConn)
}

func initDiscordClient(token string) *bot.Client {
	cacheOptions := bot.WithCacheConfigOpts(
		dgocache.WithCaches(dgocache.FlagMembers),
	)

	options := bot.WithGatewayConfigOpts(
		gateway.WithIntents(
			gateway.IntentsNonPrivileged,
			gateway.IntentGuildMembers,
			gateway.IntentsGuild,
		),
	)

	client, err := disgo.New(
		token,
		options,
		bot.WithEventListeners(commands.Handler()),
		cacheOptions,
	)
	if err != nil {
		logger.Log.Fatalf("failed to instantiate Discord client: %v", err)
	}

	return client
}

func uploadGlobalCommands(client bot.Client) {
	_, err := client.Rest.SetGlobalCommands(client.ApplicationID, commands.All())
	if err != nil {
		logger.Log.Fatalf("failed registering global commands: %v", err)
	}
	logger.Log.Info("registered global slash commands")
}

func uploadGuildCommands(client bot.Client) {
	guildId := snowflake.GetEnv("GUILD_ID")
	_, err := client.Rest.SetGuildCommands(client.ApplicationID, guildId, commands.All())
	if err != nil {
		logger.Log.Fatalf("failed registering global commands: %v", err)
	}
	logger.Log.Infof("registered guild slash commands to guild '%s'", guildId.String())
}

func waitForShutdown() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	logger.Log.Info("goodbye :)")
}
