package service

import (
	"context"
	"errors"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
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

func (m *mockGenreRepository) GetAllowedGenres(ctx context.Context) ([]genre.Genre, error) {
	return []genre.Genre{
		{ID: 1, Name: "Pop", NormalizedName: "pop"},
		{ID: 2, Name: "Rock", NormalizedName: "rock"},
		{ID: 3, Name: "Jazz", NormalizedName: "jazz"},
	}, nil
}

type mockDailyChallengeRepository struct{}

func (m *mockDailyChallengeRepository) SaveDailyChallenge(ctx context.Context, challenge *daily.Challenge) error {
	return nil
}

func (m *mockDailyChallengeRepository) SaveInverseChallenge(ctx context.Context, challenge *daily.InverseChallenge) error {
	return nil
}

type mockTimetrialChallengeRepository struct {
	savedChallenge *timetrial.Challenge
}

func (m *mockTimetrialChallengeRepository) SaveChallenge(ctx context.Context, challenge *timetrial.Challenge) error {
	m.savedChallenge = challenge
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
			country: "spain",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: 1, Name: "Pop", NormalizedName: "pop"}}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []genre.Genre{{ID: 2, Name: "Rock", NormalizedName: "rock"}}},
				{ID: "3", Name: "Song C", Artist: "Artist C", Genres: []genre.Genre{{ID: 2, Name: "Rock", NormalizedName: "rock"}}},
				{ID: "4", Name: "Song D", Artist: "Artist D", Genres: []genre.Genre{{ID: 3, Name: "Jazz", NormalizedName: "jazz"}}},
				{ID: "5", Name: "Song E", Artist: "Artist E", Genres: []genre.Genre{{ID: 1, Name: "Pop", NormalizedName: "pop"}}},
				{ID: "6", Name: "Song F", Artist: "Artist F", Genres: []genre.Genre{{ID: 2, Name: "Rock", NormalizedName: "rock"}}},
			},
			expectedError: nil,
		},
		{
			name:    "genre tie",
			country: "france",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: 1, Name: "Pop", NormalizedName: "pop"}}},
				{ID: "2", Name: "Song B", Artist: "Artist B", Genres: []genre.Genre{{ID: 2, Name: "Rock", NormalizedName: "rock"}}},
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
			country: "germany",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: 999, Name: "Unknown", NormalizedName: "unknown"}}},
			},
			expectedError: errors.New("no allowed genres found for songs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &mockMusicProvider{songs: tt.mockSongs}
			gr := &mockGenreRepository{}
			dr := &mockDailyChallengeRepository{}
			tr := &mockTimetrialChallengeRepository{}

			service := NewChallengeGenService(mp, gr, dr, tr)

			err := service.GenerateDailyChallenge(tt.country)
			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			} else if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestGenerateDailyInverseChallenge(t *testing.T) {
	tests := []struct {
		name          string
		country       string
		mockSongs     []Track
		expectedError error
	}{
		{
			name:    "valid country with songs and genres",
			country: "spain",
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: 1, Name: "Pop", NormalizedName: "pop"}}},
			},
			expectedError: nil,
		},
		{
			name:          "valid country with no songs",
			country:       "france",
			mockSongs:     []Track{},
			expectedError: errors.New("no songs found for the specified country"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &mockMusicProvider{songs: tt.mockSongs}
			gr := &mockGenreRepository{}
			dr := &mockDailyChallengeRepository{}
			tr := &mockTimetrialChallengeRepository{}

			service := NewChallengeGenService(mp, gr, dr, tr)

			err := service.GenerateInverseChallenge(tt.country)
			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			} else if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestGenerateTimetrialChallenge(t *testing.T) {
	tests := []struct {
		name          string
		countries     []string
		mockSongs     []Track
		expectedError error
	}{
		{
			name:      "generates successfully for multiple countries",
			countries: []string{"spain", "france", "italy"},
			mockSongs: []Track{
				{ID: "1", Name: "Song A", Artist: "Artist A", Genres: []genre.Genre{{ID: 1, Name: "Pop", NormalizedName: "pop"}}},
			},
			expectedError: nil,
		},
		{
			name:          "fails completely if no countries are provided",
			countries:     []string{},
			mockSongs:     []Track{},
			expectedError: errors.New("no countries provided for timetrial challenge generation"),
		},
		{
			name:          "fails and aborts if inner calculation fails for any country",
			countries:     []string{"spain", "france"},
			mockSongs:     []Track{}, // Esto forzará el error "no songs found..." en la primera iteración
			expectedError: errors.New("no songs found for the specified country"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &mockMusicProvider{songs: tt.mockSongs}
			gr := &mockGenreRepository{}
			dr := &mockDailyChallengeRepository{}
			tr := &mockTimetrialChallengeRepository{}

			service := NewChallengeGenService(mp, gr, dr, tr)

			err := service.GenerateTimetrialChallenge(tt.countries)

			// Verificar el error esperado
			if (err != nil && tt.expectedError == nil) || (err == nil && tt.expectedError != nil) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			} else if err != nil && tt.expectedError != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}

			// Si se esperaba que funcionase, verificar la estructura del desafío guardado
			if err == nil {
				if tr.savedChallenge == nil {
					t.Fatal("Expected challenge to be saved, but it was nil")
				}
				if len(tr.savedChallenge.TargetCountries) != len(tt.countries) {
					t.Errorf("Expected %d countries saved, got %d", len(tt.countries), len(tr.savedChallenge.TargetCountries))
				}
				if len(tr.savedChallenge.TargetGenres) != len(tt.countries) {
					t.Errorf("Expected %d genres saved, got %d", len(tt.countries), len(tr.savedChallenge.TargetGenres))
				}
			}
		})
	}
}
