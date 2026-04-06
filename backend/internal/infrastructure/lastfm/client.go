package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/isw2-unileon/GeoBeat/backend/internal/core/domain"
)

const baseURL = "http://ws.audioscrobbler.com/2.0/"

// Client is a simple wrapper around the Last.fm API.
type Client struct {
	APIKey string
}

func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// TODO: Map country codes to names more robustly, maybe using a library or a more comprehensive map
var codes = map[string]string{
	"US": "United States",
	"GB": "United Kingdom",
	"DE": "Germany",
	"FR": "France",
	"IT": "Italy",
	"ES": "Spain",
	"CA": "Canada",
	"AU": "Australia",
	"BR": "Brazil",
	"RU": "Russia",
}

func (c *Client) GetTopTracks(ctx context.Context, countryCode string) ([]domain.Track, error) {
	countryName, exists := codes[countryCode]
	if !exists {
		return nil, fmt.Errorf("invalid country code: %s", countryCode)
	}

	topTracksResp, err := c.getTracks(ctx, countryName)
	if err != nil {
		return nil, err
	}

	if len(topTracksResp.Tracks.Track) == 0 {
		return nil, fmt.Errorf("no tracks found for country: %s", countryCode)
	}

	tracks := []domain.Track{}
	for _, t := range topTracksResp.Tracks.Track {
		tagsResp, err := c.getTags(ctx, t.Artist.Name, t.Name)
		if err != nil {
			return nil, fmt.Errorf("error fetching tags for track %s by %s: %w", t.Name, t.Artist.Name, err)
		}

		genres := []string{}
		for _, tag := range tagsResp.Toptags.Tag {
			genres = append(genres, tag.Name)
		}

		tracks = append(tracks, domain.Track{
			Name:   t.Name,
			Artist: t.Artist.Name,
			Genres: genres,
		})
	}

	return tracks, nil
}

func (c *Client) getTracks(ctx context.Context, country string) (*TopTracksResponse, error) {
	params := url.Values{}
	params.Add("method", "geo.gettoptracks")
	params.Add("country", country)
	params.Add("api_key", c.APIKey)
	params.Add("format", "json")
	params.Add("limit", "5")

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
