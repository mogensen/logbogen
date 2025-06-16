package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WeatherService struct{}

type WeatherResponse struct {
	Daily struct {
		Temperature2mMax []float64 `json:"temperature_2m_max"`
		WeatherCode      []int     `json:"weathercode"`
	} `json:"daily"`
}

func NewWeatherService() *WeatherService {
	return &WeatherService{}
}

// GetHistoricalWeather gets the weather for a specific date and location
func (s *WeatherService) GetHistoricalWeather(lat, lng float64, date time.Time) (*WeatherResponse, error) {
	// Open-Meteo API endpoint for historical data
	url := fmt.Sprintf("https://archive-api.open-meteo.com/v1/archive?latitude=%f&longitude=%f&start_date=%s&end_date=%s&daily=temperature_2m_max,weathercode&timezone=auto",
		lat, lng, date.Format("2006-01-02"), date.Format("2006-01-02"))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned non-200 status code: %d", resp.StatusCode)
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &weatherResp, nil
}

// GetWeatherIcon returns the appropriate weather icon code based on the WMO weather code
func GetWeatherIcon(code int) string {
	// WMO Weather interpretation codes (WW)
	// https://open-meteo.com/en/docs
	switch {
	case code == 0:
		return "01d" // Clear sky
	case code >= 1 && code <= 3:
		return "02d" // Partly cloudy
	case code >= 45 && code <= 48:
		return "50d" // Fog
	case code >= 51 && code <= 55:
		return "09d" // Drizzle
	case code >= 56 && code <= 57:
		return "13d" // Freezing drizzle
	case code >= 61 && code <= 65:
		return "10d" // Rain
	case code >= 66 && code <= 67:
		return "13d" // Freezing rain
	case code >= 71 && code <= 77:
		return "13d" // Snow
	case code >= 80 && code <= 82:
		return "09d" // Rain showers
	case code >= 85 && code <= 86:
		return "13d" // Snow showers
	case code >= 95 && code <= 99:
		return "11d" // Thunderstorm
	default:
		return "01d" // Default to clear sky
	}
}
