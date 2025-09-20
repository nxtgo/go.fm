package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
	"go.fm/zlog"
)

//go:embed sql/schema.sql
var schema string

func Start(ctx context.Context, path string) (*Queries, *sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Minute)

	if _, err := sqlDB.ExecContext(ctx, schema); err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to create schema: %w", err)
	}

	queries, err := Prepare(ctx, sqlDB)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare queries: %w", err)
	}

	zlog.Log.Info("started database connection")

	return queries, sqlDB, nil
}
