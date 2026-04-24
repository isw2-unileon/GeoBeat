package service

import (
	"errors"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/isw2-unileon/GeoBeat/backend/internal/track"
)

type mockMusicProvider struct {
	songs []track.Track
}

func (m *mockMusicProvider) GetTopSongsByCountry(country string) ([]track.Track, error) {
	return m.songs, nil
}

func (m *mockMusicProvider) getSongsGenre(songs []track.Track) ([]genre.Genre, error) {
	var genres []genre.Genre

	for _, song := range songs {
		for _, stored := range m.songs {
			if song.Name == stored.Name && song.Artist == stored.Artist {
				genres = append(genres, stored.Genres...)
				break
			}
		}
	}

	return genres, nil
}

type mockGenreRepository struct{}

func (m *mockGenreRepository) GetAllowedGenres() ([]genre.Genre, error) {
	return []genre.Genre{
		{ID: "1", Name: "Pop", NormalizedName: "pop"},
		{ID: "2", Name: "Rock", NormalizedName: "rock"},
		{ID: "3", Name: "Jazz", NormalizedName: "jazz"},
	}, nil
}

type mockDailyChallengeRepository struct{}

func (m *mockDailyChallengeRepository) SaveDailyChallenge(challenge daily.Challenge) error {
	return nil
}

func TestGenerateDailyChallenge(t *testing.T) {
	tests := []struct {
		name          string
		country       string
		mockSongs     []track.Track
		expectedError error
	}{
		{
			name:    "valid country with songs and genres",
			country: "ES",
			mockSongs: []track.Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: "1", Name: "Pop", NormalizedName: "pop"}}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []genre.Genre{{ID: "2", Name: "Rock", NormalizedName: "rock"}}},
				{ID: "3", Name: "Song C", Artist: "Artist C", Genres: []genre.Genre{{ID: "3", Name: "Rock", NormalizedName: "rock"}}},
				{ID: "4", Name: "Song D", Artist: "Artist D", Genres: []genre.Genre{{ID: "4", Name: "Jazz", NormalizedName: "jazz"}}},
				{ID: "5", Name: "Song E", Artist: "Artist E", Genres: []genre.Genre{{ID: "5", Name: "Pop", NormalizedName: "pop"}}},
				{ID: "6", Name: "Song F", Artist: "Artist F", Genres: []genre.Genre{{ID: "6", Name: "Rock", NormalizedName: "rock"}}},
			},
			expectedError: nil,
		},
		{
			name:    "genre tie",
			country: "FR",
			mockSongs: []track.Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: "1", Name: "Pop", NormalizedName: "pop"}}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []genre.Genre{{ID: "2", Name: "Rock", NormalizedName: "rock"}}},
			},
			expectedError: nil,
		},
		{
			name:          "valid country with no songs",
			country:       "EmptyLand",
			mockSongs:     []track.Track{},
			expectedError: errors.New("could not fetch top songs for country: no songs found"),
		},
		{
			name:    "valid country with songs but no genres",
			country: "EN",
			mockSongs: []track.Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []genre.Genre{}},
			},
			expectedError: errors.New("could not fetch genres for songs: no genres found"),
		},
		{
			name:    "invalid genres returned by music provider",
			country: "DE",
			mockSongs: []track.Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: "999", Name: "Unknown", NormalizedName: "unknown"}}},
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
