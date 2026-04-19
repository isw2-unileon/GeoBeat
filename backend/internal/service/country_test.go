package service_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
	"github.com/isw2-unileon/GeoBeat/backend/internal/track"
)

type mockMusicProvider struct {
	tracks []track.Track
	err    error
}

func (m *mockMusicProvider) GetTopTracks(ctx context.Context, countryCode string) ([]track.Track, error) {
	return m.tracks, m.err
}

func TestCountryService_GetTopGenres(t *testing.T) {
	tests := []struct {
		name           string
		countryCode    string
		mockTracks     []track.Track
		expectedGenres []string
		expectError    bool
	}{
		{
			name:        "three genres",
			countryCode: "ES",
			mockTracks: []track.Track{
				{Genres: []string{"Pop", "Urban"}},
				{Genres: []string{"Pop", "Indie"}},
				{Genres: []string{"Pop", "Rock"}},
				{Genres: []string{"Urban"}},
				{Genres: []string{"Urban"}},
				{Genres: []string{"Urban"}},
				{Genres: []string{"Indie"}},
			},
			expectedGenres: []string{"Urban", "Pop", "Indie"},
			expectError:    false,
		},
		{
			name:        "two genres",
			countryCode: "US",
			mockTracks: []track.Track{
				{Genres: []string{"Rock"}},
				{Genres: []string{"Rock"}},
				{Genres: []string{"Country"}},
			},
			expectedGenres: []string{"Rock", "Country"},
			expectError:    false,
		},
		{
			name:           "empty genres",
			countryCode:    "FR",
			mockTracks:     []track.Track{},
			expectedGenres: []string{},
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockProvider := &mockMusicProvider{
				tracks: tc.mockTracks,
				err:    nil,
			}

			service := service.NewCountryService(mockProvider)

			country, err := service.GetCountryTopGenres(context.Background(), tc.countryCode)

			if (err != nil) != tc.expectError {
				t.Fatalf("Expected error: %v, got: %v", tc.expectError, err)
			}

			if country.Code != tc.countryCode {
				t.Errorf("Expected country %s, got %s", tc.countryCode, country.Code)
			}

			if !reflect.DeepEqual(country.TopGenres, tc.expectedGenres) {
				t.Errorf("Expected genres %v, got %v", tc.expectedGenres, country.TopGenres)
			}
		})
	}
}
