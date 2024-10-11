package operations

import (
	"context"
	"database/sql"

	"github.com/thalq/url-service/cmd/internal/logger"
	"github.com/thalq/url-service/cmd/internal/structures"
)

func QueryShortURL(ctx context.Context, db *sql.DB, shortURL string) (structures.URLData, error) {
	row := db.QueryRowContext(ctx, "SELECT original_url from urls "+
		"WHERE short_url = $1", shortURL)
	var URLData structures.URLData
	var url string
	err := row.Scan(&url)
	if err != nil {
		logger.Sugar.Errorf("Failed to get URL: %v from database", err)
		return URLData, err
	}
	logger.Sugar.Infof("Got URL: %s from database", url)
	URLData.ShortURL = shortURL
	URLData.OriginalURL = url
	return URLData, nil
}

func ExecInsertURL(ctx context.Context, db *sql.DB, URLData structures.URLData) error {
	_, err := db.ExecContext(ctx, "INSERT INTO urls (original_url, short_url) "+
		"VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING", URLData.OriginalURL, URLData.ShortURL)
	if err != nil {
		logger.Sugar.Errorf("Failed to insert URL: %v into database", err)
		return err
	}
	return nil
}
