package service

import (
	"context"
	"sort"

	"github.com/isw2-unileon/GeoBeat/backend/internal/country"
	"github.com/isw2-unileon/GeoBeat/backend/internal/track"
)

// MusicProvider defines the interface for fetching music data, allowing for different implementations (e.g., Last.fm, Spotify).
type MusicProvider interface {
	GetTopTracks(ctx context.Context, countryCode string) ([]track.Track, error)
}

// CountryService provides methods to get country-related music data, such as top genres.
type CountryService struct {
	musicProvider MusicProvider
}

// NewCountryService creates a new instance of CountryService with the given MusicProvider.
func NewCountryService(musicProvider MusicProvider) *CountryService {
	return &CountryService{
		musicProvider: musicProvider,
	}
}

// GetCountryTopGenres retrieves the top genres for a given country by analyzing the genres of the top tracks.
func (s *CountryService) GetCountryTopGenres(ctx context.Context, countryCode string) (country.Country, error) {
	tracks, err := s.musicProvider.GetTopTracks(ctx, countryCode)
	if err != nil {
		return country.Country{}, err
	}

	genreCount := make(map[string]int)
	for _, track := range tracks {
		for _, genre := range track.Genres {
			genreCount[genre]++
		}
	}

	type genreFreq struct {
		Genre string
		Freq  int
	}

	var genres []genreFreq
	for genre, freq := range genreCount {
		genres = append(genres, genreFreq{Genre: genre, Freq: freq})
	}

	sort.Slice(genres, func(i, j int) bool {
		return genres[i].Freq > genres[j].Freq
	})

	topGenres := []string{}
	for i := 0; i < len(genres) && i < 3; i++ {
		topGenres = append(topGenres, genres[i].Genre)
	}

	return country.Country{
		Code:      countryCode,
		Name:      "", // Map country code to name later
		TopGenres: topGenres,
	}, nil
}
