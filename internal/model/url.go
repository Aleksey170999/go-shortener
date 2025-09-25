package model

type URL struct {
	ID       string `json:"uuid"`
	Original string `json:"original_url"`
	Short    string `json:"short_url"`
	UserID   string `json:"user_id,omitempty"`
}

type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ShortenJSONRequest struct {
	URL string `json:"url"`
}

type ShortenJSONResponse struct {
	Result string `json:"result"`
}

type RequestURLItem struct {
	СorrelationID string `json:"correlation_id" validate:"required"`
	OriginalURL   string `json:"original_url" validate:"required"`
}

type ResponseURLItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
