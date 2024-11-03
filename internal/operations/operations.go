package operations

import (
	"context"
	"database/sql"

	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
)

func GetURLData(ctx context.Context, db *sql.DB, URL string) (models.URLData, error) {
	row := db.QueryRowContext(ctx, "SELECT original_url, short_url, correlation_id, is_deleted FROM urls "+
		"WHERE short_url = $1 OR original_url = $1", URL)
	var URLData models.URLData
	err := row.Scan(&URLData.OriginalURL, &URLData.ShortURL, &URLData.CorrelationID, &URLData.DeletedFlag)
	if err != nil {
		logger.Sugar.Errorf("Failed to get URL: %v from database", err)
		return URLData, err
	}
	logger.Sugar.Infof("Got URLData: %s from database", URLData)
	return URLData, nil
}

func GetUserURLData(ctx context.Context, db *sql.DB, userID string) ([]models.ShortURLData, error) {
	rows, err := db.QueryContext(ctx, "SELECT original_url, short_url FROM urls "+
		"WHERE user_id = $1", userID)
	if err != nil {
		logger.Sugar.Errorf("Failed to get URL: %v from database", err)
		return nil, err
	}
	defer rows.Close()
	var URLData []models.ShortURLData

	for rows.Next() {
		var data models.ShortURLData
		err := rows.Scan(&data.OriginalURL, &data.ShortURL)
		if err != nil {
			logger.Sugar.Errorf("Failed to get URL: %v from database", err)
			return nil, err
		}
		URLData = append(URLData, data)
	}

	if err = rows.Err(); err != nil {
		logger.Sugar.Errorf("Failed to iterate over rows: %v", err)
		return nil, err
	}

	logger.Sugar.Infof("Got URLData: %s from database", URLData)
	return URLData, nil
}

func InsertURL(ctx context.Context, db *sql.DB, URLData *models.URLData) error {
	_, err := db.ExecContext(ctx, "INSERT INTO urls (original_url, short_url, correlation_id, user_id) "+
		"VALUES ($1, $2, $3, $4)", URLData.OriginalURL, URLData.ShortURL, URLData.CorrelationID, URLData.UserID)
	return err
}

func ExecInsertBatchURLs(ctx context.Context, db *sql.DB, URLData []*models.URLData) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (original_url, short_url, correlation_id, user_id) "+
			"VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, data := range URLData {
		_, err := stmt.ExecContext(ctx, data.OriginalURL, data.ShortURL, data.CorrelationID, data.UserID)
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

func DeleteURLs(ctx context.Context, db *sql.DB, userId string, shortURLs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		"UPDATE urls SET is_deleted = true WHERE short_url = $1 AND user_id = $2")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, shortURL := range shortURLs {
		_, err := stmt.ExecContext(ctx, shortURL, userId)
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

func UpdateURLData(ctx context.Context, DeleteURLs <-chan models.ChDelete, results chan<- error, tx *sql.Tx) {
	for DeleteURL := range DeleteURLs {
		_, err := tx.ExecContext(ctx, "UPDATE urls SET is_deleted = true WHERE short_url = $1 AND user_id = $2", DeleteURL.ShortURL, DeleteURL.UserID)
		results <- err
		logger.Sugar.Infof("Updated URL: %s", DeleteURL.ShortURL)
	}
}
