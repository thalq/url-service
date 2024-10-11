package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/thalq/url-service/cmd/config"
)

func DBConnect(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDNS)
	if err != nil {
		return nil, err
	}
	// defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS urls (original_url TEXT PRIMARY KEY, short_url TEXT)")
	return db, nil
}
