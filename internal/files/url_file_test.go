package files

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thalq/url-service/config"
	"github.com/thalq/url-service/internal/models"
)

func TestProducer_WriteEvent(t *testing.T) {
	cfg := config.Config{FileStoragePath: "test_data.log"}
	producer, err := NewProducer(cfg.FileStoragePath)
	assert.NoError(t, err)
	defer producer.Close()

	urlData := &models.URLData{
		OriginalURL:   "http://example.com",
		ShortURL:      "exmpl",
		CorrelationID: "12345",
		UserID:        "user1",
	}

	err = producer.WriteEvent(urlData)
	assert.NoError(t, err)

	file, err := os.Open(cfg.FileStoragePath)
	assert.NoError(t, err)
	defer file.Close()

	var readData models.URLData
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err = json.Unmarshal(scanner.Bytes(), &readData)
		assert.NoError(t, err)
		assert.Equal(t, urlData, &readData)
	}
	assert.NoError(t, scanner.Err())

	err = os.Remove(cfg.FileStoragePath)
	assert.NoError(t, err)
}

func TestConsumer_GetURL(t *testing.T) {
	cfg := config.Config{FileStoragePath: "test_data.log"}
	producer, err := NewProducer(cfg.FileStoragePath)
	assert.NoError(t, err)
	defer producer.Close()

	urlData := &models.URLData{
		OriginalURL:   "http://example.com",
		ShortURL:      "exmpl",
		CorrelationID: "12345",
		UserID:        "user1",
	}

	err = producer.WriteEvent(urlData)
	assert.NoError(t, err)
	producer.Close()

	consumer, err := NewConsumer(cfg.FileStoragePath)
	assert.NoError(t, err)
	defer consumer.Close()

	originalURL, err := consumer.GetURL("exmpl")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", originalURL)

	// Удаляем тестовый файл
	err = os.Remove(cfg.FileStoragePath)
	assert.NoError(t, err)
}

func TestConsumer_GetURLsByUser(t *testing.T) {
	cfg := config.Config{FileStoragePath: "test_data.log"}
	producer, err := NewProducer(cfg.FileStoragePath)
	assert.NoError(t, err)
	defer producer.Close()

	urlData1 := &models.URLData{
		OriginalURL:   "http://example1.com",
		ShortURL:      "exmpl1",
		CorrelationID: "12345",
		UserID:        "user1",
	}
	urlData2 := &models.URLData{
		OriginalURL:   "http://example2.com",
		ShortURL:      "exmpl2",
		CorrelationID: "12346",
		UserID:        "user1",
	}

	err = producer.WriteEvent(urlData1)
	assert.NoError(t, err)
	err = producer.WriteEvent(urlData2)
	assert.NoError(t, err)
	producer.Close()

	consumer, err := NewConsumer(cfg.FileStoragePath)
	assert.NoError(t, err)
	defer consumer.Close()

	urls, err := consumer.GetURLsByUser("user1")
	assert.NoError(t, err)
	assert.Len(t, urls, 2)
	assert.Equal(t, urlData1, urls[0])
	assert.Equal(t, urlData2, urls[1])

	err = os.Remove(cfg.FileStoragePath)
	assert.NoError(t, err)
}

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
