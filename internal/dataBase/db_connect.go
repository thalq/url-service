package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/thalq/url-service/config"
)

func DBConnect(cfg config.Config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDNS)

	if err != nil {
		return nil
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS urls (original_url TEXT PRIMARY KEY, short_url TEXT, correlation_id TEXT, user_id TEXT, is_deleted BOOL DEFAULT False)")

	if err != nil {
		return nil
	}
	return db
}
