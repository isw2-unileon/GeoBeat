package domain

// Country represents a country with its code, name, and top music genres.
type Country struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	TopGenres []string `json:"top_genres"`
}
