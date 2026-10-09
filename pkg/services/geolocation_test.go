package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stubGeocoder struct {
	response *GeocodeResponse
	err      error
}

func (s stubGeocoder) ReverseGeocode(lat, lng float64) (*GeocodeResponse, error) {
	return s.response, s.err
}

func TestActivityService_resolveLocation(t *testing.T) {
	tests := []struct {
		name      string
		geocoder  Geocoder
		submitted string
		want      string
	}{
		{name: "uses the geocoded name", geocoder: stubGeocoder{response: &GeocodeResponse{Name: "Amager Strand"}}, submitted: "x", want: "Amager Strand"},
		{name: "falls back to the submitted location on error", geocoder: stubGeocoder{err: errors.New("status 429")}, submitted: " Havnen ", want: "Havnen"},
		{name: "falls back to coordinates on error", geocoder: stubGeocoder{err: errors.New("timeout")}, want: "55.6761, 12.5683"},
		{name: "falls back when nothing is found", geocoder: stubGeocoder{response: &GeocodeResponse{}}, submitted: "Havnen", want: "Havnen"},
		{name: "works without a geocoder", geocoder: nil, want: "55.6761, 12.5683"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := ActivityService{geocoder: tt.geocoder}
			require.Equal(t, tt.want, sut.resolveLocation(55.676098, 12.568337, tt.submitted))
		})
	}
}

func TestNominatimGeocoder_ReverseGeocode(t *testing.T) {
	newGeocoder := func(handler http.HandlerFunc) *NominatimGeocoder {
		server := httptest.NewServer(handler)
		t.Cleanup(server.Close)
		return &NominatimGeocoder{baseURL: server.URL, client: server.Client()}
	}

	t.Run("parses the response and sends the user agent", func(t *testing.T) {
		geocoder := newGeocoder(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/reverse", r.URL.Path)
			require.Equal(t, nominatimUserAgent, r.Header.Get("User-Agent"))
			require.Equal(t, "55.676098", r.URL.Query().Get("lat"))
			_, _ = w.Write([]byte(`{"name":"Amager Strand","address":{"country":"Danmark"}}`))
		})

		got, err := geocoder.ReverseGeocode(55.676098, 12.568337)
		require.NoError(t, err)
		require.Equal(t, "Amager Strand, Danmark", got.SimpleDisplayName())
	})

	t.Run("rate limiting is an error", func(t *testing.T) {
		geocoder := newGeocoder(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		})

		got, err := geocoder.ReverseGeocode(55, 12)
		require.ErrorContains(t, err, "status 429")
		require.Nil(t, got)
	})

	t.Run("invalid json is an error", func(t *testing.T) {
		geocoder := newGeocoder(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`<html>`))
		})

		_, err := geocoder.ReverseGeocode(55, 12)
		require.Error(t, err)
	})

	t.Run("a slow server times out", func(t *testing.T) {
		geocoder := newGeocoder(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
		})
		geocoder.client = &http.Client{Timeout: 20 * time.Millisecond}

		_, err := geocoder.ReverseGeocode(55, 12)
		require.Error(t, err)
	})

	t.Run("the default client has a timeout", func(t *testing.T) {
		require.NotZero(t, NewNominatimGeocoder().client.Timeout)
	})
}
