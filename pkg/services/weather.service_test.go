package services

import (
	"testing"
	"time"
)

func TestGetWeather(t *testing.T) {
	// Create a new weather service
	service := NewWeatherService()

	// Test case: Copenhagen, Denmark on a specific date
	// Using coordinates for Copenhagen
	lat := 55.676098
	lng := 12.568337
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	// Get weather data
	weather, err := service.GetWeather(lat, lng, date)
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

	// Verify temperature at that date is 11.9°C
	temp := weather.Daily.Temperature2mMax[0]
	if temp < 11 || temp > 12 {
		t.Errorf("Temperature %.1f°C should be around 11.9°C", temp)
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

func TestGetWeatherToday(t *testing.T) {
	// Create a new weather service
	service := NewWeatherService()

	// Test case: North Pole
	lat := 90.0 // North Pole latitude
	lng := 0.0  // North Pole longitude
	date := time.Now()

	// Get weather data
	weather, err := service.GetWeather(lat, lng, date)
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

	// Verify temperature is within reasonable range for North Pole
	// North Pole temperatures typically range from -40°C to 0°C
	temp := weather.Daily.Temperature2mMax[0]
	if temp > 10 {
		t.Errorf("Temperature %.1f°C is above 0°C at the North Pole", temp)
	}

	// Verify weather code is valid (0-99)
	code := weather.Daily.WeatherCode[0]
	if code < 0 || code > 99 {
		t.Errorf("Weather code %d is outside valid range (0-99)", code)
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
