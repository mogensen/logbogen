package utils

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/types"
)

// ParseBody is helper function for parsing the body.
// Is any error occurs it will panic.
// Its just a helper function to avoid writing if condition again n again.
func ParseBody(ctx *fiber.Ctx, body interface{}) *fiber.Error {
	if err := ctx.BodyParser(body); err != nil {
		slog.Error("Error parsing body", "error", err)
		return fiber.ErrBadRequest
	}

	return nil
}

// ParseBodyAndValidate is helper function for parsing the body.
// Is any error occurs it will panic.
// Its just a helper function to avoid writing if condition again n again.
func ParseBodyAndValidate(ctx *fiber.Ctx, body interface{}) *fiber.Error {
	if err := ParseBody(ctx, body); err != nil {
		return err
	}

	return Validate(body)
}

// GetUser is helper function for getting authenticated user's id
func GetUser(c *fiber.Ctx) *uint {
	id, _ := c.Locals("USER").(uint)
	return &id
}

func IsCurrentUser(c *fiber.Ctx, userId uint) bool {
	currentUserId := GetUser(c)
	if currentUserId == nil {

		return false
	}
	if *currentUserId == userId {
		return true
	}
	return false
}

func GetCsrf(c *fiber.Ctx) string {
	csrf, _ := c.Locals("csrf").(string)
	return csrf
}

func FormatDate(date types.Date) string {
	return time.Time(date).Format("2006-01-02")
}

func IsSameUser(a, b *uint) bool {
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ToJSON(v []types.ClimbingActivity) (template.HTML, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return template.HTML(string(b)), nil
}
