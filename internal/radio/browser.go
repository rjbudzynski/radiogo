package radio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/rjbudzynski/radiogo/internal/config"
)

// Station represents a radio station from the Radio Browser API.
type Station struct {
	UUID     string `json:"stationuuid"`
	Name     string `json:"name"`
	URL      string `json:"url_resolved"`
	Homepage string `json:"homepage"`
	Tags     string `json:"tags"`
	Country  string `json:"country"`
	Language string `json:"language"`
	Codec    string `json:"codec"`
	Bitrate  int    `json:"bitrate"`
	Votes    int    `json:"votes"`
	Favicon  string `json:"favicon"`
}

// Category represents a metadata group (tag, country, or language).
type Category struct {
	Name  string `json:"name"`
	Count int    `json:"stationcount"`
}

// SearchOptions describes a Radio Browser station search.
type SearchOptions struct {
	Name       string
	Country    string
	Language   string
	Tag        string
	Codec      string
	Order      string
	Reverse    bool
	Limit      int
	BitrateMin int
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func apiGet(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, config.APIBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", config.APIUserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}
	return json.Unmarshal(body, out)
}

// TopStations returns the most-voted stations.
func TopStations(limit int) ([]Station, error) {
	return SearchStationsWithOptions(SearchOptions{
		Order:   "votes",
		Reverse: true,
		Limit:   limit,
	})
}

// SearchStations searches by name, sorted by votes descending.
func SearchStations(query string, limit int) ([]Station, error) {
	return SearchStationsWithOptions(SearchOptions{
		Name:    query,
		Order:   "votes",
		Reverse: true,
		Limit:   limit,
	})
}

// SearchStationsWithOptions performs a station search with optional filters and sorting.
func SearchStationsWithOptions(opts SearchOptions) ([]Station, error) {
	path := buildSearchPath(opts)
	var stations []Station
	err := apiGet(path, &stations)
	return stations, err
}

func buildSearchPath(opts SearchOptions) string {
	values := url.Values{}
	if opts.Name != "" {
		values.Set("name", opts.Name)
	}
	if opts.Country != "" {
		values.Set("country", opts.Country)
	}
	if opts.Language != "" {
		values.Set("language", opts.Language)
	}
	if opts.Tag != "" {
		values.Set("tag", opts.Tag)
	}
	if opts.Codec != "" {
		values.Set("codec", opts.Codec)
	}
	if opts.BitrateMin > 0 {
		values.Set("bitrateMin", strconv.Itoa(opts.BitrateMin))
	}
	if opts.Order != "" {
		values.Set("order", opts.Order)
	}
	if opts.Reverse {
		values.Set("reverse", "true")
	}
	if opts.Limit > 0 {
		values.Set("limit", strconv.Itoa(opts.Limit))
	}
	values.Set("hidebroken", "true")
	return "/stations/search?" + values.Encode()
}

// ListTags returns tags with at least 100 stations, sorted by station count descending.
func ListTags(limit int) ([]Category, error) {
	var cats []Category
	err := apiGet(fmt.Sprintf("/tags?limit=%d&order=stationcount&reverse=true&hidebroken=true", limit), &cats)
	return cats, err
}

// ListCountries returns countries, sorted by station count descending.
func ListCountries() ([]Category, error) {
	var cats []Category
	err := apiGet("/countries?order=stationcount&reverse=true&hidebroken=true", &cats)
	return cats, err
}

// ListLanguages returns languages, sorted by station count descending.
func ListLanguages() ([]Category, error) {
	var cats []Category
	err := apiGet("/languages?order=stationcount&reverse=true&hidebroken=true", &cats)
	return cats, err
}

// ListCodecs returns codecs sorted by station count descending.
func ListCodecs(limit int) ([]Category, error) {
	var cats []Category
	err := apiGet(fmt.Sprintf("/codecs?limit=%d&order=stationcount&reverse=true&hidebroken=true", limit), &cats)
	return cats, err
}

// SearchByCategory searches for stations with a specific tag, country, language, or codec.
func SearchByCategory(filterType, value string, limit int) ([]Station, error) {
	opts := SearchOptions{Limit: limit, Order: "votes", Reverse: true}
	switch filterType {
	case "tag":
		opts.Tag = value
	case "country":
		opts.Country = value
	case "language":
		opts.Language = value
	case "codec":
		opts.Codec = value
	}
	return SearchStationsWithOptions(opts)
}
