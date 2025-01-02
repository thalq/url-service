package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/thalq/url-service/config"
	logger "github.com/thalq/url-service/internal/middleware"
)

func DBConnect(cfg config.Config) *sql.DB {
	db, err := sql.Open("pgx", cfg.DatabaseDNS)

	if err != nil {
		logger.Sugar.Error("Failed to connect to database")
		return nil
	}
	logger.Sugar.Info("Connected to database")

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS urls (original_url TEXT PRIMARY KEY, short_url TEXT, correlation_id TEXT, user_id TEXT, is_deleted BOOL DEFAULT False)")

	if err != nil {
		logger.Sugar.Error("Failed to create table")
		return nil
	}
	logger.Sugar.Info("Table created")
	return db
}
