package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitList(t *testing.T) {
	require.Empty(t, splitList(""))
	require.Equal(t, []string{"127.0.0.1"}, splitList("127.0.0.1"))
	require.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, splitList(" 10.0.0.1, ,10.0.0.2 "))
}

func TestDefaultDB(t *testing.T) {
	require.Equal(t, "./data/logbog.db", DefaultDB)
	require.Equal(t, DefaultDB, getEnv("DB_NOT_SET_IN_TESTS", DefaultDB))
}
