package carbon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/viral32111/carbon-intensity-exporter/internal/iso8601"
)

const (
	REGIONAL_API_URL = "https://api.carbonintensity.org.uk/regional"
)

type RegionalData struct {
	From iso8601.ISO8601 `json:"from"` // ISO-8601
	To   iso8601.ISO8601 `json:"to"`   // ISO-8601

	Regions []struct {
		ID       int    `json:"regionid"`
		Name     string `json:"shortname"`
		Operator string `json:"dnoregion"`

		Intensity struct {
			Forecast int    `json:"forecast"` // gCO2/kWh
			Index    string `json:"index"`
		} `json:"intensity"`

		GenerationMix []struct {
			Fuel       string  `json:"fuel"`
			Percentage float64 `json:"perc"`
		} `json:"generationmix"`
	} `json:"regions"`
}

type APIResponse struct {
	Data []RegionalData `json:"data"`
}

func FetchRegionalData() ([]RegionalData, error) {
	httpClient := &http.Client{Timeout: 5 * time.Second}

	httpResponse, err := httpClient.Get(REGIONAL_API_URL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch regional data from '%s': %w", REGIONAL_API_URL, err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received unsuccessful HTTP status code %d when fetching regional data from '%s'", httpResponse.StatusCode, REGIONAL_API_URL)
	}

	var apiResponse APIResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("Failed to parse regional data from '%s': %w", REGIONAL_API_URL, err)
	}

	return apiResponse.Data, nil
}
