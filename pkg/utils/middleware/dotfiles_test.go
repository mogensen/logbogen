package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestBlockDotfiles(t *testing.T) {
	app := fiber.New()
	app.Use(BlockDotfiles)
	app.Get("/*", func(c *fiber.Ctx) error { return c.SendString("ok") })

	tests := []struct {
		path string
		want int
	}{
		{"/js/application.js", fiber.StatusOK},
		{"/js/v1.2/app.js", fiber.StatusOK},
		{"/js/.DS_Store", fiber.StatusNotFound},
		{"/.env", fiber.StatusNotFound},
		{"/.git/config", fiber.StatusNotFound},
		{"/js/%2eDS_Store", fiber.StatusNotFound},
		{"/css/%2e%2e/.env", fiber.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", tt.path, nil))
			require.NoError(t, err)
			require.Equal(t, tt.want, resp.StatusCode)
		})
	}
}
