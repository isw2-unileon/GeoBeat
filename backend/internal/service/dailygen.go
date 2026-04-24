package service

import (
	"context"
	"errors"
	"time"

	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/isw2-unileon/GeoBeat/backend/internal/track"
)

type MusicProvider interface {
	GetTopSongsByCountry(ctx context.Context, country string) ([]track.Track, error)
	GetSongsGenre(ctx context.Context, songs []track.Track) ([]string, error)
}

type GenreRepository interface {
	GetAllowedGenres(ctx context.Context) ([]genre.Genre, error)
}

type DailyChallengeRepository interface {
	SaveDailyChallenge(ctx context.Context, challenge daily.Challenge) error
}

type DailyChallengeService struct {
	musicProvider MusicProvider
	genreRepo     GenreRepository
	dailyRepo     DailyChallengeRepository
}

func NewDailyChallengeService(mp MusicProvider, gr GenreRepository, dr DailyChallengeRepository) *DailyChallengeService {
	return &DailyChallengeService{
		musicProvider: mp,
		genreRepo:     gr,
		dailyRepo:     dr,
	}
}

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
	for _, g := range genres {
		if _, ok := allowedGenreSet[g]; ok {
			genreCount[g]++
		}
	}

	if len(genreCount) == 0 {
		return errors.New("no allowed genres found for songs")
	}

	// TODO add hint songs for the genre

	var topGenre string
	maxCount := 0
	for g, count := range genreCount {
		if count > maxCount {
			topGenre = g
			maxCount = count
		}
	}

	challenge := daily.Challenge{
		TargetCountry: country,
		TargetGenre:   topGenre,
		Date:          time.Now().Truncate(24 * time.Hour),
	}

	return s.dailyRepo.SaveDailyChallenge(ctx, challenge)
}
