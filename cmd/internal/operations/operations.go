package operations

import (
	"context"
	"database/sql"

	"net/http"

	"github.com/thalq/url-service/cmd/config"
	"github.com/thalq/url-service/cmd/internal/files"
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

func ExecInsertURL(ctx context.Context, db *sql.DB, URLData *structures.URLData) error {
	_, err := db.ExecContext(ctx, "INSERT INTO urls (original_url, short_url) "+
		"VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING", URLData.OriginalURL, URLData.ShortURL)
	if err != nil {
		logger.Sugar.Errorf("Failed to insert URL: %v into database", err)
		return err
	}
	return nil
}

func ExecInsertBatchURLs(ctx context.Context, db *sql.DB, URLData []structures.URLData) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (original_url, short_url) "+
			"VALUES ($1, $2) ON CONFLICT (original_url) DO NOTHING")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, data := range URLData {
		_, err := stmt.ExecContext(ctx, data.OriginalURL, data.ShortURL)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func InserDataIntoDB(ctx context.Context, db *sql.DB, URLData *structures.URLData) error {
	logger.Sugar.Infoln("Database connection established")
	if err := ExecInsertURL(ctx, db, URLData); err != nil {
		logger.Sugar.Fatal(err)
	}
	logger.Sugar.Infof("URL inserted into database: %s:%s", URLData.OriginalURL, URLData.ShortURL)
	return nil
}

func InsertDataIntoFile(cfg config.Config, URLData *structures.URLData) error {
	Producer, err := files.NewProducer(cfg.FileStoragePath)
	if err != nil {
		logger.Sugar.Fatal(err)
	}
	defer Producer.Close()
	if err := Producer.WriteEvent(URLData); err != nil {
		logger.Sugar.Fatal(err)
	}
	logger.Sugar.Infof("URL inserted into file: %s:%s", URLData.OriginalURL, URLData.ShortURL)
	return nil
}

func GetOriginalURLFromDB(db *sql.DB, shortURL string, w http.ResponseWriter, r *http.Request) {
	logger.Sugar.Infoln("Database connection established")
	URLData, err := QueryShortURL(r.Context(), db, shortURL)
	if err != nil {
		http.Error(w, "ShortURL not found in database", http.StatusNotFound)
		return
	}
	logger.Sugar.Infoln("GET: Original URL from database:", URLData.OriginalURL)
	w.Header().Set("Location", URLData.OriginalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
	logger.Sugar.Infoln("Temporary Redirect sent for URL:", URLData.OriginalURL)
}
