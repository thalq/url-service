package structures

type URLData struct {
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}
