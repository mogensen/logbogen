package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWeatherDescription(t *testing.T) {
	tests := map[int]string{0: "Klart", 2: "Halvskyet", 46: "Tåge", 53: "Støvregn", 63: "Regn", 73: "Sne", 96: "Torden", 10: ""}
	for code, want := range tests {
		require.Equal(t, want, Weather{Code: code}.Description(), "code %d", code)
	}
}
