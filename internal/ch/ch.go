package ch

import (
	"context"
	"database/sql"
	"sync"

	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
	"github.com/thalq/url-service/internal/operations"
)

func FanIn(ctx context.Context, db *sql.DB, DeleteURLs <-chan models.ChDelete, tx *sql.Tx) <-chan error {
	results := make(chan error)
	var wg sync.WaitGroup
	const numWorkers = 3

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for deleteURL := range DeleteURLs {
				err := operations.UpdateURLData(ctx, deleteURL, tx)
				results <- err
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

func Generate(urlsToDelete ...models.ChDelete) chan models.ChDelete {
	outCh := make(chan models.ChDelete)
	go func() {
		for _, url := range urlsToDelete {
			outCh <- url
		}
		close(outCh)
	}()
	return outCh
}

func DeleteURLData(ctx context.Context, db *sql.DB, UrlsToDelete ...models.ChDelete) {
	tx, err := db.Begin()
	if err != nil {
		logger.Sugar.Fatalf("Failed to start transaction: %v", err)
	}
	deleteURLs := Generate(UrlsToDelete...)
	results := FanIn(ctx, db, deleteURLs, tx)

	for err := range results {
		if err != nil {
			logger.Sugar.Error("Failed to update URL:", err)
			tx.Rollback()
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Sugar.Error("Failed to commit transaction:", err)
	}
}
