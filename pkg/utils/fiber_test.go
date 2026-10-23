package utils

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

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

func TestNormalizeNewlines(t *testing.T) {
	type inner struct{ Note string }
	type form struct {
		Comment string
		Role    string
		Count   int
		Nested  *inner
		hidden  string
	}

	value := &form{Comment: "a\r\nb\r\n", Role: "no breaks", Count: 3, Nested: &inner{Note: "x\r\ny"}, hidden: "h\r\ni"}

	normalizeNewlines(reflect.ValueOf(value))

	require.Equal(t, "a\nb\n", value.Comment)
	require.Equal(t, "no breaks", value.Role)
	require.Equal(t, 3, value.Count)
	require.Equal(t, "x\ny", value.Nested.Note, "nested structs are covered")
	require.Equal(t, "h\r\ni", value.hidden, "unexported fields are left alone")
	require.NotPanics(t, func() { normalizeNewlines(reflect.ValueOf((*form)(nil))) })
}

func TestValidate_MessagesNameTheFieldAndTheLimitButNotTheValue(t *testing.T) {
	long := strings.Repeat("hemmelig", 400)
	activity := &types.Activity{
		CategoryID: "climbing", TypeID: "boulder", Comment: long, Role: strings.Repeat("r", 101),
		Lat: 95, ParticipantsIDs: make([]uint64, 51),
	}

	err := Validate(activity)

	require.NotNil(t, err)
	require.Equal(t, 400, err.Code)
	require.Contains(t, err.Message, "Kommentaren må højst være 2000 tegn lang.")
	require.Contains(t, err.Message, "Rollen må højst være 100 tegn lang.")
	require.Contains(t, err.Message, "Der kan højst vælges 50 deltagere.")
	require.Contains(t, err.Message, "Positionen på kortet er ugyldig.")
	require.NotContains(t, err.Message, "hemmelig")
	require.NotContains(t, err.Message, "constraint")

	require.Nil(t, Validate(&types.Activity{CategoryID: "climbing", TypeID: "boulder"}))
}

func TestFormatDateLong(t *testing.T) {
	day := types.Date(time.Date(2026, time.October, 3, 0, 0, 0, 0, time.UTC))
	require.Equal(t, "3. oktober 2026", FormatDateLong(day))
}
