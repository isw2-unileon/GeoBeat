package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/isw2-unileon/GeoBeat/backend/internal/timetrial"
)

// Track represents a music track with its associated genres.
type Track struct {
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Artist string        `json:"artist"`
	Genres []genre.Genre `json:"genres"`
}

// MusicProvider defines the interface for fetching music data.
type MusicProvider interface {
	GetTopSongsByCountry(ctx context.Context, country string) ([]Track, error)
	GetSongsGenre(ctx context.Context, songs []Track) ([][]string, error)
}

// DailyChallengeRepository defines the interface for saving daily challenges.
type DailyChallengeRepository interface {
	SaveDailyChallenge(ctx context.Context, challenge *daily.Challenge) error
	SaveInverseChallenge(ctx context.Context, challenge *daily.InverseChallenge) error
}

// TimetrialChallengeSaverRepository defines the interface for saving time trial challenges.
type TimetrialChallengeSaverRepository interface {
	SaveChallenge(ctx context.Context, c *timetrial.Challenge) error
}

// GenreRepository defines the interface for managing genres.
type GenreRepository interface {
	GetAllowedGenres(ctx context.Context) ([]genre.Genre, error)
}

// ChallengeGenService is responsible for generating and saving the daily and timetrial challenges.
type ChallengeGenService struct {
	musicProvider MusicProvider
	genreRepo     GenreRepository
	dailyRepo     DailyChallengeRepository
	timetrialRepo TimetrialChallengeSaverRepository
}

// NewChallengeGenService creates a new instance of ChallengeGenService with the provided dependencies.
func NewChallengeGenService(mp MusicProvider, gr GenreRepository, dr DailyChallengeRepository, tr TimetrialChallengeSaverRepository) *ChallengeGenService {
	return &ChallengeGenService{
		musicProvider: mp,
		genreRepo:     gr,
		dailyRepo:     dr,
		timetrialRepo: tr,
	}
}

// CalculateTopGenreAndHints calculates the top genre and hint songs for a given country. It returns the top genre, a list of hint songs, and any error encountered during the process.
func (s *ChallengeGenService) CalculateTopGenreAndHints(ctx context.Context, country string) (string, []string, error) {
	songs, err := s.musicProvider.GetTopSongsByCountry(ctx, country)
	if err != nil {
		return "", nil, err
	}

	if len(songs) == 0 {
		return "", nil, errors.New("no songs found for the specified country")
	}

	genres, err := s.musicProvider.GetSongsGenre(ctx, songs)
	if err != nil {
		return "", nil, err
	}

	allowedGenres, err := s.genreRepo.GetAllowedGenres(ctx)
	if err != nil {
		return "", nil, err
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
		return "", nil, errors.New("no allowed genres found for songs")
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

	return topGenre, hintSongs, nil
}

// GenerateDailyChallenge generates a new daily challenge based on the top songs and genres of a specified country and saves it to the repository.
func (s *ChallengeGenService) GenerateDailyChallenge(country string) error {
	ctx := context.Background()

	topGenre, hintSongs, err := s.CalculateTopGenreAndHints(ctx, country)
	if err != nil {
		return err
	}

	challenge := daily.Challenge{
		TargetCountry: country,
		TargetGenre:   topGenre,
		HintSongs:     hintSongs,
		Date:          time.Now().Truncate(24 * time.Hour),
	}

	return s.dailyRepo.SaveDailyChallenge(ctx, &challenge)
}

// GenerateTimetrialChallenge generates a new time trial challenge based on the top genres of the specified countries and saves it to the repository.
func (s *ChallengeGenService) GenerateTimetrialChallenge(countries []string) error {
	ctx := context.Background()

	if len(countries) == 0 {
		return errors.New("no countries provided for timetrial challenge generation")
	}

	targetGenres := make([]string, 0, len(countries))

	for _, country := range countries {
		topGenre, _, err := s.CalculateTopGenreAndHints(ctx, country)
		if err != nil {
			return err
		}
		targetGenres = append(targetGenres, topGenre)
	}

	challenge := timetrial.Challenge{
		TargetCountries: countries,
		TargetGenres:    targetGenres,
		Date:            time.Now().Truncate(24 * time.Hour),
	}

	return s.timetrialRepo.SaveChallenge(ctx, &challenge)
}

// GenerateInverseChallenge generates a new inverse daily challenge based on the top song of a specified country and saves it to the repository.
func (s *ChallengeGenService) GenerateInverseChallenge(country string) error {
	ctx := context.Background()
	songs, err := s.musicProvider.GetTopSongsByCountry(ctx, country)
	if err != nil {
		return err
	}

	if len(songs) == 0 {
		return errors.New("no songs found for the specified country")
	}

	challenge := daily.InverseChallenge{
		GivenSongName: songs[0].Name,
		TargetCountry: country,
		Date:          time.Now().Truncate(24 * time.Hour),
	}

	return s.dailyRepo.SaveInverseChallenge(ctx, &challenge)
}
