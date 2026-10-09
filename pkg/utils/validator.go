package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// fieldLabels name the fields a user can fill in, in the words of the forms
var fieldLabels = map[string]string{
	"Comment":    "Kommentaren",
	"Role":       "Rollen",
	"OtherType":  "Den anden type",
	"Location":   "Lokationen",
	"Provider":   "Udbyderen",
	"CategoryID": "Kategorien",
	"TypeID":     "Typen",
}

// Validate validates the input struct. The message names the field and the limit but never the submitted
// value, and the technical details go to the log.
func Validate(payload interface{}) *fiber.Error {
	err := validate.Struct(payload)
	if err == nil {
		return nil
	}

	var invalid validator.ValidationErrors
	if !errors.As(err, &invalid) {
		slog.Error("Validation error", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Ugyldige oplysninger")
	}

	var details, messages []string
	for _, fieldErr := range invalid {
		details = append(details, fmt.Sprintf("%v:%v", fieldErr.Field(), fieldErr.Tag()))
		messages = append(messages, validationMessage(fieldErr))
	}
	slog.Warn("Validation failed", "fields", strings.Join(details, ","))
	return fiber.NewError(fiber.StatusBadRequest, strings.Join(messages, " "))
}

// validationMessage explains one failed rule to the user
func validationMessage(fieldErr validator.FieldError) string {
	label, known := fieldLabels[fieldErr.Field()]

	switch fieldErr.Tag() {
	case "max":
		if fieldErr.Kind() == reflect.Slice {
			return fmt.Sprintf("Der kan højst vælges %s deltagere.", fieldErr.Param())
		}
		if known {
			return fmt.Sprintf("%s må højst være %s tegn lang.", label, fieldErr.Param())
		}
	case "required", "required_if", "required_unless":
		if known {
			return fmt.Sprintf("%s skal udfyldes.", label)
		}
	case "gte", "lte":
		if fieldErr.Field() == "Lat" || fieldErr.Field() == "Lng" {
			return "Positionen på kortet er ugyldig."
		}
	}
	return "Ugyldige oplysninger."
}

// CUSTOM VALIDATION RULES =============================================

// Password validation rule: required,min=6,max=100
var _ = validate.RegisterValidation("password", func(fl validator.FieldLevel) bool {
	l := len(fl.Field().String())

	return l >= 6 && l < 100
})
