package operations

import (
	"context"
	"database/sql"

	"github.com/thalq/url-service/config"
	"github.com/thalq/url-service/internal/files"
	"github.com/thalq/url-service/internal/logger"
	"github.com/thalq/url-service/internal/structures"
)

func GetURLData(ctx context.Context, db *sql.DB, URL string) (structures.URLData, error) {
	row := db.QueryRowContext(ctx, "SELECT original_url, short_url, correlation_id FROM urls "+
		"WHERE short_url = $1 OR original_url = $1", URL)
	var URLData structures.URLData
	err := row.Scan(&URLData.OriginalURL, &URLData.ShortURL, &URLData.CorrelationID)
	if err != nil {
		logger.Sugar.Errorf("Failed to get URL: %v from database", err)
		return URLData, err
	}
	logger.Sugar.Infof("Got URLData: %s from database", URLData)
	return URLData, nil
}

func InsertURL(ctx context.Context, db *sql.DB, URLData *structures.URLData) error {
	_, err := db.ExecContext(ctx, "INSERT INTO urls (original_url, short_url, correlation_id) "+
		"VALUES ($1, $2, $3)", URLData.OriginalURL, URLData.ShortURL, URLData.CorrelationID)
	return err
}

func ExecInsertBatchURLs(ctx context.Context, db *sql.DB, URLData []*structures.URLData) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (original_url, short_url, correlation_id) "+
			"VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, data := range URLData {
		_, err := stmt.ExecContext(ctx, data.OriginalURL, data.ShortURL, data.CorrelationID)
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

func InsertDataIntoFile(cfg config.Config, URLData *structures.URLData) error {
	Producer, err := files.NewProducer(cfg.FileStoragePath)
	if err != nil {
		logger.Sugar.Error(err)
	}
	defer Producer.Close()
	if err := Producer.WriteEvent(URLData); err != nil {
		logger.Sugar.Error(err)
	}
	logger.Sugar.Infof("URL inserted into file: %s:%s", URLData.OriginalURL, URLData.ShortURL)
	return nil
}

func InsertBatchIntoFile(cfg config.Config, URLData []*structures.URLData) error {
	Producer, err := files.NewProducer(cfg.FileStoragePath)
	if err != nil {
		logger.Sugar.Error(err)
	}
	defer Producer.Close()
	for _, data := range URLData {
		toFileSaveData := &structures.URLData{
			CorrelationID: data.CorrelationID,
			OriginalURL:   data.OriginalURL,
			ShortURL:      data.ShortURL,
		}
		if err := Producer.WriteEvent(toFileSaveData); err != nil {
			logger.Sugar.Errorf("Failed to write data to file: %v", err)
		}
	}
	logger.Sugar.Infoln("Data saved to file")
	return nil
}
