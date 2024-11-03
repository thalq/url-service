package models

import "github.com/golang-jwt/jwt/v4"

type URLData struct {
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
	UserID        string `json:"user_id"`
	DeletedFlag   bool   `db:"is_deleted"`
}

type ShortURLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchURLRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchURLResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type DeleteRequest struct {
	ShortURLs []string `json:"short_urls"`
}

type ChDelete struct {
	UserID   string `json:"user_id"`
	ShortURL string `json:"short_url"`
}
