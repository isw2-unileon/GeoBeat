package domain

type Track struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Artist string   `json:"artist"`
	Genres []string `json:"genres"`
}
