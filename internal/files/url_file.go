package files

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/thalq/url-service/internal/structures"
)

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteEvent(URLData *structures.URLData) error {
	data, err := json.Marshal(&URLData)
	if err != nil {
		return err
	}
	// добавляем перенос строки
	if _, err := p.writer.Write(data); err != nil {
		return err
	}
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}
	return p.writer.Flush()
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) GetURL(shortURL string) (string, error) {
	for c.scanner.Scan() {
		var data structures.URLData
		if err := json.Unmarshal(c.scanner.Bytes(), &data); err != nil {
			return "", err
		}

		if data.ShortURL == shortURL {
			return data.OriginalURL, nil
		}
	}

	if err := c.scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("ShortURL not found")
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
