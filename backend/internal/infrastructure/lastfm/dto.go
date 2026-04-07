package lastfm

// TopTracksResponse is the DTO for parsing the top tracks response for a country
type TopTracksResponse struct {
	Tracks struct {
		Track []struct {
			Name   string `json:"name"`
			Artist struct {
				Name string `json:"name"`
			} `json:"artist"`
		} `json:"track"`
	} `json:"tracks"`
}

// TopTagsResponse is the DTO for parsing the top tags response for a track
type TopTagsResponse struct {
	Toptags struct {
		Tag []struct {
			Name string `json:"name"`
		} `json:"tag"`
	} `json:"toptags"`
}
