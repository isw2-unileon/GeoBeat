package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
)

// Genre represents a music genre.
type Genre struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
}

// Track represents a music track with its associated genres.
type Track struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Artist string  `json:"artist"`
	Genres []Genre `json:"genres"`
}

// MusicProvider defines the interface for fetching music data.
type MusicProvider interface {
	GetTopSongsByCountry(ctx context.Context, country string) ([]Track, error)
	GetSongsGenre(ctx context.Context, songs []Track) ([][]string, error)
}

// GenreRepository defines the interface for fetching allowed genres.
type GenreRepository interface {
	GetAllowedGenres(ctx context.Context) ([]Genre, error)
}

// DailyChallengeRepository defines the interface for saving daily challenges.
type DailyChallengeRepository interface {
	SaveDailyChallenge(ctx context.Context, challenge daily.Challenge) error
}

// DailyChallengeService is responsible for generating and saving the daily challenge.
type DailyChallengeService struct {
	musicProvider MusicProvider
	genreRepo     GenreRepository
	dailyRepo     DailyChallengeRepository
}

// NewDailyChallengeService creates a new instance of DailyChallengeService with the provided dependencies.
func NewDailyChallengeService(mp MusicProvider, gr GenreRepository, dr DailyChallengeRepository) *DailyChallengeService {
	return &DailyChallengeService{
		musicProvider: mp,
		genreRepo:     gr,
		dailyRepo:     dr,
	}
}

// GenerateDailyChallenge generates a new daily challenge based on the top songs and genres of a specified country and saves it to the repository.
func (s *DailyChallengeService) GenerateDailyChallenge(country string) error {
	ctx := context.Background()
	songs, err := s.musicProvider.GetTopSongsByCountry(ctx, country)
	if err != nil {
		return err
	}

	if len(songs) == 0 {
		return errors.New("no songs found for the specified country")
	}

	genres, err := s.musicProvider.GetSongsGenre(ctx, songs)
	if err != nil {
		return err
	}

	allowedGenres, err := s.genreRepo.GetAllowedGenres(ctx)
	if err != nil {
		return err
	}

	allowedGenreSet := make(map[string]struct{})
	for _, g := range allowedGenres {
		allowedGenreSet[g.NormalizedName] = struct{}{}
	}

	genreCount := make(map[string]int)
	for _, songGenres := range genres {
		for _, genre := range songGenres {
			if _, ok := allowedGenreSet[genre]; ok {
				genreCount[genre]++
			}
		}
	}

	if len(genreCount) == 0 {
		return errors.New("no allowed genres found for songs")
	}

	var topGenre string
	maxCount := 0
	for g, count := range genreCount {
		if count > maxCount {
			topGenre = g
			maxCount = count
		}
	}

	var hintSongs []string
	for i, songGenres := range genres {
		if slices.Contains(songGenres, topGenre) {
			hintSongs = append(hintSongs, songs[i].Name)
		}
	}

	challenge := daily.Challenge{
		TargetCountry: country,
		TargetGenre:   topGenre,
		HintSongs:     hintSongs,
		Date:          time.Now().Truncate(24 * time.Hour),
	}

	return s.dailyRepo.SaveDailyChallenge(ctx, challenge)
}
