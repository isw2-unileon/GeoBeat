package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

type mockMusicProvider struct {
	songs []Track
}

func (m *mockMusicProvider) GetTopSongsByCountry(ctx context.Context, country string) ([]Track, error) {
	return m.songs, nil
}

func (m *mockMusicProvider) GetSongsGenre(ctx context.Context, songs []Track) ([][]string, error) {
	var genres [][]string

	for _, song := range songs {
		var songGenres []string
		for _, genre := range song.Genres {
			songGenres = append(songGenres, genre.NormalizedName)
		}
		genres = append(genres, songGenres)
	}

	return genres, nil
}

type mockGenreRepository struct{}

func (m *mockGenreRepository) GetAllowedGenres(ctx context.Context) ([]Genre, error) {
	return []Genre{
		{ID: "1", Name: "Pop", NormalizedName: "pop"},
		{ID: "2", Name: "Rock", NormalizedName: "rock"},
		{ID: "3", Name: "Jazz", NormalizedName: "jazz"},
	}, nil
}

type mockDailyChallengeRepository struct{}

func (m *mockDailyChallengeRepository) SaveDailyChallenge(ctx context.Context, challenge daily.Challenge) error {
	return nil
}

func TestGenerateDailyChallenge(t *testing.T) {
	tests := []struct {
		name          string
		country       string
		mockSongs     []Track
		expectedError error
	}{
		{
			name:    "valid country with songs and genres",
			country: "ES",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []Genre{{ID: "1", Name: "Pop", NormalizedName: "pop"}}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []Genre{{ID: "2", Name: "Rock", NormalizedName: "rock"}}},
				{ID: "3", Name: "Song C", Artist: "Artist C", Genres: []Genre{{ID: "2", Name: "Rock", NormalizedName: "rock"}}},
				{ID: "4", Name: "Song D", Artist: "Artist D", Genres: []Genre{{ID: "3", Name: "Jazz", NormalizedName: "jazz"}}},
				{ID: "5", Name: "Song E", Artist: "Artist E", Genres: []Genre{{ID: "1", Name: "Pop", NormalizedName: "pop"}}},
				{ID: "6", Name: "Song F", Artist: "Artist F", Genres: []Genre{{ID: "2", Name: "Rock", NormalizedName: "rock"}}},
			},
			expectedError: nil,
		},
		{
			name:    "genre tie",
			country: "FR",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []Genre{{ID: "1", Name: "Pop", NormalizedName: "pop"}}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []Genre{{ID: "2", Name: "Rock", NormalizedName: "rock"}}},
			},
			expectedError: nil,
		},
		{
			name:          "valid country with no songs",
			country:       "EmptyLand",
			mockSongs:     []Track{},
			expectedError: errors.New("no songs found for the specified country"),
		},
		{
			name:    "invalid genres returned by music provider",
			country: "DE",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []Genre{{ID: "999", Name: "Unknown", NormalizedName: "unknown"}}},
			},
			expectedError: errors.New("no allowed genres found for songs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &mockMusicProvider{songs: tt.mockSongs}
			gr := &mockGenreRepository{}
			dr := &mockDailyChallengeRepository{}

			service := NewDailyChallengeService(mp, gr, dr)

			err := service.GenerateDailyChallenge(tt.country)
			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			} else if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}
