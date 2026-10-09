package utils

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// genericErrorMessage is shown for every error that is not a deliberate client error
const genericErrorMessage = "Der opstod en fejl."

// ErrorHandler is used to catch error thrown inside the routes by ctx.Next(err).
// Only the message of a 4xx fiber error is shown, anything else is logged and hidden.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := genericErrorMessage

	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
		if code < fiber.StatusInternalServerError {
			msg = fe.Message
		}
	}

	if code >= fiber.StatusInternalServerError {
		slog.Error("Request failed", "method", c.Method(), "path", c.Path(), "status", code, "error", err)
	}

	c.Status(code)
	return c.Render("errors/error", fiber.Map{
		"Error": msg,
		"Code":  code,
	})
}

// DBError logs a data layer error and turns it into a fiber error without internal details
func DBError(err error) *fiber.Error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "Not found")
	}
	slog.Error("Database error", "error", err)
	return fiber.NewError(fiber.StatusInternalServerError, genericErrorMessage)
}
