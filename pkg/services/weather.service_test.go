package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTestWeatherService serves the handler for both the forecast and the archive endpoint
func newTestWeatherService(t *testing.T, handler http.HandlerFunc) *WeatherService {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return NewWeatherServiceWithBaseURLs(server.URL+"/forecast", server.URL+"/archive")
}

func TestGetWeather_Historical(t *testing.T) {
	service := newTestWeatherService(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/archive", r.URL.Path)
		require.Equal(t, "55.676098", r.URL.Query().Get("latitude"))
		require.Equal(t, "12.568337", r.URL.Query().Get("longitude"))
		require.Equal(t, "2024-03-15", r.URL.Query().Get("start_date"))
		require.Equal(t, "2024-03-15", r.URL.Query().Get("end_date"))
		_, _ = w.Write([]byte(`{"daily":{"temperature_2m_max":[11.9],"weathercode":[53]}}`))
	})

	weather, err := service.GetWeather(55.676098, 12.568337, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	temperature, code, ok := weather.Day()
	require.True(t, ok)
	require.Equal(t, 11.9, temperature)
	require.Equal(t, 53, code)
	require.Equal(t, "09d", GetWeatherIcon(code))
}

func TestGetWeather_Today(t *testing.T) {
	service := newTestWeatherService(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/forecast", r.URL.Path)
		require.Equal(t, "90.000000", r.URL.Query().Get("latitude"))
		_, _ = w.Write([]byte(`{"current":{"temperature_2m":-21.5,"weathercode":3}}`))
	})

	weather, err := service.GetWeather(90, 0, time.Now())
	require.NoError(t, err)
	// The current weather is returned in the same shape as the daily weather
	temperature, code, ok := weather.Day()
	require.True(t, ok)
	require.Equal(t, -21.5, temperature)
	require.Equal(t, 3, code)
}

func TestGetWeather_Failures(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name:    "non-200 status",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTooManyRequests) },
			wantErr: "non-200 status code: 429",
		},
		{
			name:    "invalid json",
			handler: func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("<html>")) },
			wantErr: "failed to decode",
		},
	}

	for _, tt := range tests {
		for _, day := range []time.Time{date, time.Now()} {
			t.Run(tt.name+" "+day.Format("2006-01-02"), func(t *testing.T) {
				weather, err := newTestWeatherService(t, tt.handler).GetWeather(55, 12, day)
				require.ErrorContains(t, err, tt.wantErr)
				require.Nil(t, weather)
			})
		}
	}
}

func TestGetWeather_SlowServerTimesOut(t *testing.T) {
	service := newTestWeatherService(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	})
	service.client.Timeout = 30 * time.Millisecond

	_, err := service.GetWeather(55, 12, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))
	require.Error(t, err)
}

func TestNewWeatherService_HasTimeoutAndPublicEndpoints(t *testing.T) {
	service := NewWeatherService()

	require.Equal(t, 5*time.Second, service.client.Timeout)
	require.Equal(t, "https://api.open-meteo.com/v1/forecast", service.forecastURL)
	require.Equal(t, "https://archive-api.open-meteo.com/v1/archive", service.archiveURL)
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

func TestWeatherResponse_Day(t *testing.T) {
	decode := func(t *testing.T, body string) *WeatherResponse {
		var response WeatherResponse
		require.NoError(t, json.Unmarshal([]byte(body), &response))
		return &response
	}

	t.Run("values are returned", func(t *testing.T) {
		temperature, code, ok := decode(t, `{"daily":{"temperature_2m_max":[0],"weathercode":[0]}}`).Day()
		require.True(t, ok, "a real 0 degrees and clear sky is valid")
		require.Zero(t, temperature)
		require.Zero(t, code)
	})

	t.Run("null values are missing, not zero", func(t *testing.T) {
		_, _, ok := decode(t, `{"daily":{"temperature_2m_max":[null],"weathercode":[null]}}`).Day()
		require.False(t, ok)
		_, _, ok = decode(t, `{"daily":{"temperature_2m_max":[4.5],"weathercode":[null]}}`).Day()
		require.False(t, ok)
	})

	t.Run("empty and nil responses are missing", func(t *testing.T) {
		_, _, ok := decode(t, `{"daily":{}}`).Day()
		require.False(t, ok)
		_, _, ok = (*WeatherResponse)(nil).Day()
		require.False(t, ok)
	})
}
