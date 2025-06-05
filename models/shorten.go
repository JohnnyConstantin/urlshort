package models

// ShortenRequest Объект, содержащий полный URL
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse Объект, содержащий сокращенный URL
type ShortenResponse struct {
	Result string `json:"result"`
}

// URLResponse Объект, содержащий сокращенный URL и соответствующий ему полный URL
type URLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchShortenRequest В дальнейшем возможно будет использован для группировки полных URL под одним ID
// Пока что бесполезен
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortenResponse В дальнейшем возможно будет использован для группировки сокращенных URL под одним ID.
// Пока что бесполезен
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
