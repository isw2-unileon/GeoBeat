package service

import (
	"github.com/isw2-unileon/GeoBeat/backend/internal/daily"
	"github.com/isw2-unileon/GeoBeat/backend/internal/genre"
	"github.com/isw2-unileon/GeoBeat/backend/internal/track"
)

type MusicProvider interface {
	GetTopSongsByCountry(country string) ([]track.Track, error)
	getSongsGenre(songs []track.Track) ([]genre.Genre, error)
}

type GenreRepository interface {
	GetAllowedGenres() ([]genre.Genre, error)
}

type DailyChallengeRepository interface {
	SaveDailyChallenge(challenge daily.Challenge) error
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
	return nil
}
