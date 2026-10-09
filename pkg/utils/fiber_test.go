package utils

import (
	"net/url"
	"strings"
	"testing"

	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestUserImage_NameStaysInOnePathSegment(t *testing.T) {
	names := []string{"Jane Doe", "a/b", "../../etc", "x?y=1#z", "<script>alert(1)</script>", "100%"}

	for _, name := range names {
		got := UserImage(types.User{Name: name, Email: "jane@example.com"})

		gravatar, err := url.Parse(got)
		require.NoError(t, err, name)
		require.Equal(t, "www.gravatar.com", gravatar.Host, name)

		fallback, err := url.Parse(gravatar.Query().Get("d"))
		require.NoError(t, err, name)
		require.Equal(t, "ui-avatars.com", fallback.Host, name)
		require.Empty(t, fallback.RawQuery, name)
		require.Empty(t, fallback.Fragment, name)

		// /api/<name>/256/<color> must keep exactly four segments whatever the name contains
		require.Len(t, strings.Split(fallback.EscapedPath(), "/"), 5, name)
		require.NotContains(t, got, "<", name)
	}
}
