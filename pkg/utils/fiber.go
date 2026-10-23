package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/ruang-guru/monday"
)

var _emptyDate = types.Date{}

// ParseBody is helper function for parsing the body.
// Is any error occurs it will panic.
// Its just a helper function to avoid writing if condition again n again.
func ParseBody(ctx *fiber.Ctx, body interface{}) *fiber.Error {
	if err := ctx.BodyParser(body); err != nil {
		Log(ctx).Warn("Error parsing body", "error", err)
		return fiber.ErrBadRequest
	}

	normalizeNewlines(reflect.ValueOf(body))
	return nil
}

// normalizeNewlines turns the CRLF line breaks of a submitted textarea into LF in every string field. Browsers
// count a break as one character in maxlength but submit two, so text that fits the form could fail validation.
func normalizeNewlines(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			normalizeNewlines(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				normalizeNewlines(v.Field(i))
			}
		}
	case reflect.String:
		if v.CanSet() {
			v.SetString(strings.ReplaceAll(v.String(), "\r\n", "\n"))
		}
	}
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
func GetUser(c *fiber.Ctx) *types.User {
	user, _ := c.Locals("USER").(*types.User)
	return user
}

func IsCurrentUser(c *fiber.Ctx, userId uint64) bool {
	currentUser := GetUser(c)
	if currentUser == nil {

		return false
	}
	if currentUser.ID == userId {
		return true
	}
	return false
}

func FormatDate(date types.Date) string {
	day := time.Time(date)
	if date == _emptyDate {
		day = time.Now()
	}
	return time.Time(day).Format("2006-01-02")
}

func FormatDateHuman(date types.Date) string {
	day := time.Time(date)
	if date == _emptyDate {
		day = time.Now()
	}

	res := monday.Format(day, "Monday 2 January 2006", monday.LocaleDaDK)
	// uppercase the first letter
	res = string(res[0]-32) + res[1:]
	return res
}

// FormatDateLong formats a date as "23. oktober 2026"
func FormatDateLong(date types.Date) string {
	return monday.Format(time.Time(date), "2. January 2006", monday.LocaleDaDK)
}

func IsSameUser(a, b *uint64) bool {
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func FirstSix(s []types.User) []types.User {
	if len(s) > 6 {
		return s[:6]
	}
	return s
}

func UserImage(user types.User) string {
	colors := []string{"32ab5a", "efdc9d", "70b74a", "97c650", "ebce81", "807a4e", "c4c69b", "439133", "6d6d5b", "d9d9d3"}
	hash := 0
	for _, char := range user.Name {
		hash += int(char)
	}
	color := colors[hash%len(colors)]

	fallBack := fmt.Sprintf("https://ui-avatars.com/api/%s/256/%s", url.PathEscape(user.Name), color)
	encoded := url.QueryEscape(fallBack)
	hasher := md5.Sum([]byte(strings.TrimSpace(user.Email)))
	emailHash := hex.EncodeToString(hasher[:])

	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=%s&s=200", emailHash, encoded)
}

type ActivityCtx struct {
	UserID   *uint64
	Activity *types.Activity
}

func CtxActivity(user uint64, activity types.Activity) ActivityCtx {
	return ActivityCtx{
		UserID:   &user,
		Activity: &activity,
	}
}
