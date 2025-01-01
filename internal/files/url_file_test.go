package files

import (
	"os"
	"testing"

	"github.com/thalq/url-service/config"
	"github.com/thalq/url-service/internal/models"
)

func BenchmarkWriteEvent(b *testing.B) {
	cfg := config.Config{FileStoragePath: "test_data.log"}
	producer, err := NewProducer(cfg.FileStoragePath)
	if err != nil {
		b.Fatalf("Failed to create producer: %v", err)
	}
	consumer, err := NewConsumer(cfg.FileStoragePath)
	if err != nil {
		b.Fatalf("Failed to create consumer: %v", err)
	}
	defer producer.Close()
	defer consumer.Close()

	urlData := &models.URLData{
		OriginalURL:   "http://example.com",
		ShortURL:      "exmpl",
		CorrelationID: "12345",
		UserID:        "user1",
	}

	b.ResetTimer()

	b.Run("write", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := producer.WriteEvent(urlData); err != nil {
				b.Fatalf("Failed to write event: %v", err)
			}
		}
	})
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := consumer.GetURL("exmpl")
			if err != nil && err.Error() != "ShortURL not found" {
				b.Fatalf("Failed to get URL: %v", err)
			}
		}
	})
	b.Run("get by user", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := consumer.GetURLsByUser("user1")
			if err != nil {
				b.Fatalf("Failed to get URLs by user: %v", err)
			}
		}
	})

	err = os.Remove(cfg.FileStoragePath)
	if err != nil && !os.IsNotExist(err) {
		b.Fatalf("Failed to remove file: %v", err)
	}
}
