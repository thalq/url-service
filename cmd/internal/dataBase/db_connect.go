package database

import (
	"context"
	"database/sql"

	"github.com/thalq/url-service/cmd/config"
)

func DBConnect(cfg config.Config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDNS)
	if err != nil {
		return nil
	}
	// defer db.Close()
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return nil
	}
	db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS urls (original_url TEXT PRIMARY KEY, short_url TEXT)")
	return db
}
