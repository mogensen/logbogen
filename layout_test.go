package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Google Fonts CSS differs per browser, so it cannot carry an integrity hash
const fontsHost = "https://fonts.googleapis.com/"

func TestLayoutExternalAssetsHaveIntegrity(t *testing.T) {
	html, err := os.ReadFile("views/layouts/main.html")
	require.NoError(t, err)

	tags := regexp.MustCompile(`<script[^>]*src="https://[^>]*>|<link[^>]*rel="stylesheet"[^>]*href="https://[^>]*>`).
		FindAllString(string(html), -1)
	require.NotEmpty(t, tags)

	for _, tag := range tags {
		if strings.Contains(tag, fontsHost) {
			continue
		}
		require.Contains(t, tag, `integrity="sha`, "external asset without integrity: %s", tag)
		require.Contains(t, tag, "crossorigin", "external asset without crossorigin: %s", tag)
	}
}

func TestLayoutExternalAssetsArePinned(t *testing.T) {
	html, err := os.ReadFile("views/layouts/main.html")
	require.NoError(t, err)

	urls := regexp.MustCompile(`(?:src|href)="(https://(?:unpkg\.com|cdn\.jsdelivr\.net/npm)/[^"]+)"`).FindAllStringSubmatch(string(html), -1)
	require.NotEmpty(t, urls)

	for _, m := range urls {
		require.Regexp(t, `@\d+\.\d+\.\d+`, m[1], "asset version is not pinned: %s", m[1])
	}
}
