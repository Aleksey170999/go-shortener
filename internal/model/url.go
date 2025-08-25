package model

type URL struct {
	ID       string `json:"uuid"`
	Original string `json:"original_url"`
	Short    string `json:"short_url"`
}

type ShortenJSONRequest struct {
	URL string `json:"url"`
}

type ShortenJSONResponse struct {
	Result string `json:"result"`
}
