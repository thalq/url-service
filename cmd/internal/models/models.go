package models

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchURLRequest struct {
	// CorrelationID string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type BatchURLResponse struct {
	// CorrelationID string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}
