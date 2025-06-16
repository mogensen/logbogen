package services

import (
	"testing"
	"time"
)

func TestGetHistoricalWeather(t *testing.T) {
	// Create a new weather service
	service := NewWeatherService()

	// Test case: Copenhagen, Denmark on a specific date
	// Using coordinates for Copenhagen
	lat := 55.676098
	lng := 12.568337
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	// Get weather data
	weather, err := service.GetHistoricalWeather(lat, lng, date)
	if err != nil {
		t.Fatalf("Failed to get weather data: %v", err)
	}

	// Verify response structure
	if len(weather.Daily.Temperature2mMax) == 0 {
		t.Error("Expected temperature data, got empty array")
	}
	if len(weather.Daily.WeatherCode) == 0 {
		t.Error("Expected weather code data, got empty array")
	}

	// Verify temperature is within reasonable range for Copenhagen in March
	// Temperature at that date is 11.9°C
	temp := weather.Daily.Temperature2mMax[0]
	if temp < 11 || temp > 12 {
		t.Errorf("Temperature %.1f°C is outside expected range for Copenhagen in March", temp)
	}

	// Verify weather code is 53 (partly cloudy) for that date
	code := weather.Daily.WeatherCode[0]
	if code != 53 {
		t.Errorf("Weather code %d should be 53", code)
	}

	// Test weather icon mapping
	icon := GetWeatherIcon(code)
	if icon == "" {
		t.Error("Expected weather icon code, got empty string")
	}
}

func TestGetWeatherIcon(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"Clear sky", 0, "01d"},
		{"Partly cloudy", 1, "02d"},
		{"Fog", 45, "50d"},
		{"Drizzle", 51, "09d"},
		{"Freezing drizzle", 56, "13d"},
		{"Rain", 61, "10d"},
		{"Freezing rain", 66, "13d"},
		{"Snow", 71, "13d"},
		{"Rain showers", 80, "09d"},
		{"Snow showers", 85, "13d"},
		{"Thunderstorm", 95, "11d"},
		{"Invalid code", 999, "01d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWeatherIcon(tt.code)
			if got != tt.expected {
				t.Errorf("GetWeatherIcon(%d) = %v, want %v", tt.code, got, tt.expected)
			}
		})
	}
}
