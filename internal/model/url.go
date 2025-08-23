package model

type URL struct {
	ID       string
	Original string
}

type ShortenJSONRequest struct {
	URL string `json:"url"`
}

type ShortenJSONResponse struct {
	Result string `json:"result"`
}
