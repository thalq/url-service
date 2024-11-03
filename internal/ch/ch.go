package ch

import (
	"context"
	"database/sql"
	"sync"

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
