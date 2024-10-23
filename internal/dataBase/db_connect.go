package database

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/thalq/url-service/config"
)

func DBConnect(cfg config.Config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDNS)
	if err != nil {
		return nil
	}
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return nil
	}
	db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS urls (original_url TEXT PRIMARY KEY, short_url TEXT, correlation_id TEXT, user_id TEXT)")
	return db
}
