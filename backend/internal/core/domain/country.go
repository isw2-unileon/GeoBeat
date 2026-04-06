package domain

type Country struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	TopGenres []string `json:"top_genres"`
}
