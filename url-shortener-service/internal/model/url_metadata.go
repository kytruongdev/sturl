package model

type UrlMetadata struct {
	FinalURL    string `json:"final_url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Favicon     string `json:"favicon"`
}

func (u UrlMetadata) IsNotEmpty() bool {
	return u.FinalURL != "" || u.Title != "" || u.Description != "" || u.Image != "" || u.Favicon != ""
}
