package services

import (
	"context"
	"sort"

	"github.com/isw2-unileon/GeoBeat/backend/internal/core/domain"
)

type MusicProvider interface {
	GetTopTracks(ctx context.Context, countryCode string) ([]domain.Track, error)
}

type CountryService struct {
	musicProvider MusicProvider
}

func NewCountryService(musicProvider MusicProvider) *CountryService {
	return &CountryService{
		musicProvider: musicProvider,
	}
}

func (s *CountryService) GetCountryTopGenres(ctx context.Context, countryCode string) (domain.Country, error) {
	tracks, err := s.musicProvider.GetTopTracks(ctx, countryCode)
	if err != nil {
		return domain.Country{}, err
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

	return domain.Country{
		Code:      countryCode,
		Name:      "", // Map country code to name later
		TopGenres: topGenres,
	}, nil
}
