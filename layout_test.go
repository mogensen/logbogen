package main

import (
	"io/fs"
	"os"
	"path/filepath"
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

// The CSP has no unsafe-inline for scripts, so templates must not carry inline scripts or event handlers
func TestTemplatesHaveNoInlineScripts(t *testing.T) {
	script := regexp.MustCompile(`(?s)<script([^>]*)>(.*?)</script>`)
	handler := regexp.MustCompile(`\son[a-z]+\s*=\s*["']|href\s*=\s*["']javascript:`)

	err := filepath.WalkDir("views", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".html" {
			return err
		}
		content, err := os.ReadFile(path)
		require.NoError(t, err)

		for _, match := range script.FindAllStringSubmatch(string(content), -1) {
			require.Contains(t, match[1], "src=", "inline script in %s", path)
			require.Empty(t, strings.TrimSpace(match[2]), "script with src has a body in %s", path)
		}
		require.False(t, handler.MatchString(string(content)), "inline event handler in %s", path)
		return nil
	})
	require.NoError(t, err)
}
