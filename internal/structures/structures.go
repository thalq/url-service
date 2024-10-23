package structures

import "github.com/golang-jwt/jwt/v4"

type URLData struct {
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
	UserID        string `json:"user_id"`
}

type ShortURLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}
