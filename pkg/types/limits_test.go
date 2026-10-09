package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaxLength(t *testing.T) {
	t.Run("reads the max from the validate tag", func(t *testing.T) {
		got, err := MaxLength(Activity{}, "Comment")
		require.NoError(t, err)
		require.Equal(t, 2000, got)
	})

	t.Run("finds it among other rules", func(t *testing.T) {
		got, err := MaxLength(Activity{}, "OtherType")
		require.NoError(t, err)
		require.Equal(t, 100, got)
	})

	t.Run("fields without a max are an error, not zero", func(t *testing.T) {
		_, err := MaxLength(Activity{}, "Lat")
		require.ErrorContains(t, err, "no max")
	})

	t.Run("unknown fields are an error", func(t *testing.T) {
		_, err := MaxLength(Activity{}, "Nope")
		require.ErrorContains(t, err, "no field Nope")
	})
}

func TestTheFormsHaveALimitForEveryTextField(t *testing.T) {
	for _, field := range []string{"Role", "Comment", "OtherType", "Location", "ParticipantsIDs"} {
		limit, err := ActivityMax(field)
		require.NoError(t, err, field)
		require.Positive(t, limit, field)
	}
	for _, field := range []string{"Provider", "OtherType", "ParticipantsIDs"} {
		limit, err := CertificationMax(field)
		require.NoError(t, err, field)
		require.Positive(t, limit, field)
	}
}
