package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// Validate validates the input struct. The details only go to the log, since they echo field names and values.
func Validate(payload interface{}) *fiber.Error {
	err := validate.Struct(payload)
	if err != nil {
		var invalid validator.ValidationErrors
		if errors.As(err, &invalid) {
			var details []string
			for _, fieldErr := range invalid {
				details = append(details, fmt.Sprintf("%v:%v", fieldErr.Field(), fieldErr.Tag()))
			}
			slog.Warn("Validation failed", "fields", strings.Join(details, ","))
		} else {
			slog.Error("Validation error", "error", err)
		}

		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	return nil
}

// CUSTOM VALIDATION RULES =============================================

// Password validation rule: required,min=6,max=100
var _ = validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
	l := len(fl.Field().String())

	return l >= 6 && l < 100
})
