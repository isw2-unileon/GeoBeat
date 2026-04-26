package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/isw2-unileon/GeoBeat/backend/internal/service"
)

// TODO extract api url to env variable
const baseURL = "http://ws.audioscrobbler.com/2.0/"

// Client is a simple wrapper around the Last.fm API.
type Client struct {
	APIKey string
}

// NewClient creates a new Last.fm API client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// GetTopTracksByCountry fetches the top tracks for a given country code from the Last.fm API.
func (c *Client) GetTopTracksByCountry(ctx context.Context, countryCode string) ([]service.Track, error) {
	resp, err := c.getTracks(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("error fetching top tracks for country %s: %w", countryCode, err)
	}

	var tracks []service.Track
	for _, t := range resp.Tracks.Track {
		tracks = append(tracks, service.Track{
			Name:   t.Name,
			Artist: t.Artist.Name,
		})
	}
	return tracks, nil
}

// GetSongsGenre fetches the genres for a list of songs using the Last.fm API.
func (c *Client) GetSongsGenre(ctx context.Context, songs []service.Track) ([][]string, error) {
	var allGenres [][]string
	for _, song := range songs {
		resp, err := c.getTags(ctx, song.Artist, song.Name)
		if err != nil {
			return nil, fmt.Errorf("error fetching tags for song %s by artist %s: %w", song.Name, song.Artist, err)
		}

		var genres []string
		for _, tag := range resp.Toptags.Tag {
			// Remove any non-alphanumeric characters and trim whitespace to get a clean genre name
			formattedTag := strings.Map(func(r rune) rune {
				if unicode.IsLetter(r) || unicode.IsDigit(r) {
					return r
				}
				return -1
			}, strings.TrimSpace(tag.Name))
			genres = append(genres, formattedTag)
		}
		allGenres = append(allGenres, genres)

		// To avoid hitting rate limits, we could add a small delay here if needed
	}
	return allGenres, nil
}

func (c *Client) getTracks(ctx context.Context, country string) (*TopTracksResponse, error) {
	params := url.Values{}
	params.Add("method", "geo.gettoptracks")
	params.Add("country", country)
	params.Add("api_key", c.APIKey)
	params.Add("format", "json")
	params.Add("limit", "50")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to last.fm (GetTopTracks): %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to last.fm (GetTopTracks): %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last.fm returned an unexpected status: %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result TopTracksResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("error decoding JSON (GetTopTracks): %w", err)
	}

	return &result, nil
}

func (c *Client) getTags(ctx context.Context, artistName string, trackName string) (*TopTagsResponse, error) {
	params := url.Values{}
	params.Add("method", "track.gettoptags")
	params.Add("artist", artistName)
	params.Add("track", trackName)
	params.Add("api_key", c.APIKey)
	params.Add("format", "json")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to last.fm (GetTopTags): %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to last.fm (GetTopTags): %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last.fm returned an unexpected status: %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result TopTagsResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("error decoding JSON (GetTopTags): %w", err)
	}

	return &result, nil
}
