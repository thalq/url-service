package ch

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	logger "github.com/thalq/url-service/internal/middleware"
	"github.com/thalq/url-service/internal/models"
)

func TestFanIn(t *testing.T) {
	logger.InitLogger()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)

	deleteURLs := make(chan models.ChDelete, 3)
	deleteURLs <- models.ChDelete{ShortURL: "short1", UserID: "user1"}
	deleteURLs <- models.ChDelete{ShortURL: "short2", UserID: "user2"}
	deleteURLs <- models.ChDelete{ShortURL: "short3", UserID: "user3"}
	close(deleteURLs)

	// Ожидаемым порядке, в котором будут обработки из deleteURLs
	mock.ExpectExec("UPDATE urls SET is_deleted = true WHERE short_url = \\$1 AND user_id = \\$2").
		WithArgs("short1", "user1").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE urls SET is_deleted = true WHERE short_url = \\$1 AND user_id = \\$2").
		WithArgs("short2", "user2").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE urls SET is_deleted = true WHERE short_url = \\$1 AND user_id = \\$2").
		WithArgs("short3", "user3").WillReturnResult(sqlmock.NewResult(1, 1))

	results := FanIn(ctx, db, deleteURLs, tx)

	// Соберите все ошибки
	for err := range results {
		if err != nil {
			t.Errorf("Error occurred: %v", err)
		}
	}

	mock.ExpectCommit()
	err = tx.Commit()
	if err != nil {
		t.Errorf("Error during commit: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestGenerate(t *testing.T) {
	urlsToDelete := []models.ChDelete{
		{ShortURL: "short1", UserID: "user1"},
		{ShortURL: "short2", UserID: "user2"},
		{ShortURL: "short3", UserID: "user3"},
	}

	outCh := Generate(urlsToDelete...)
	var results []models.ChDelete
	for url := range outCh {
		results = append(results, url)
	}

	assert.Equal(t, urlsToDelete, results)
}
