package radio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/woodz-dot/radiogo/internal/config"
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
	var stations []Station
	err := apiGet(fmt.Sprintf("/stations/topvote/%d", limit), &stations)
	return stations, err
}

// SearchStations searches by name, sorted by votes descending.
func SearchStations(query string, limit int) ([]Station, error) {
	q := url.QueryEscape(query)
	path := fmt.Sprintf("/stations/search?name=%s&limit=%d&order=votes&reverse=true&hidebroken=true", q, limit)
	var stations []Station
	err := apiGet(path, &stations)
	return stations, err
}
