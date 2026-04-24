package track

import "github.com/isw2-unileon/GeoBeat/backend/internal/genre"

// Track represents a music track with its ID, name, artist, and associated genres.
type Track struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Artist string        `json:"artist"`
	Genres []genre.Genre `json:"genres"`
}
