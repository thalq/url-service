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
			operations.UpdateURLData(ctx, DeleteURLs, results, tx)
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
		defer close(outCh)
		for _, n := range urlsToDelete {
			outCh <- n
		}
	}()

	return outCh
}

func DeleteURLData(ctx context.Context, db *sql.DB, UrlsToDelete ...models.ChDelete) {
	tx, err := db.Begin()
	if err != nil {
		logger.Sugar.Fatalf("Failed to start transaction: %v", err)
	}
	userChan := Generate(UrlsToDelete...)
	results := FanIn(ctx, db, userChan, tx)

	hasErrors := false
	for err := range results {
		if err != nil {
			hasErrors = true
			logger.Sugar.Infof("Error occurred: %v", err)
		}
	}

	if hasErrors {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Sugar.Infof("tx rollback error: %v", rollbackErr)
		}
		logger.Sugar.Infof("Transaction rolled back due to errors")
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			logger.Sugar.Infof("tx commit error: %v", commitErr)
		}
		logger.Sugar.Infoln("Transaction committed successfully")
	}
}
