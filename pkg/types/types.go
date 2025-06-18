package types

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
)

// MsgResponse defined the message payload
type MsgResponse struct {
	Message string `json:"message"`
}

type Date time.Time

func (ct Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(ct).Format("2006-01-02") + `"`), nil
}

func (ct *Date) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*ct = Date(t)
	return nil
}
func (ct *Date) Time() time.Time {
	return time.Time(*ct)
}

// String() returns the time in string
func (ct *Date) String() string {
	t := time.Time(*ct).String()
	return t
}

// Register the converter for CustomTime type format as 2006-01-02
var timeConverter = func(value string) reflect.Value {
	slog.Debug("Converting time value", "value", value)
	if v, err := time.Parse("2006-01-02", value); err == nil {
		return reflect.ValueOf(v)
	}
	return reflect.Value{}
}

func init() {
	customTime := fiber.ParserType{
		Customtype: Date{},
		Converter:  timeConverter,
	}

	// Add setting to the Decoder
	fiber.SetParserDecoder(fiber.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType:        []fiber.ParserType{customTime},
		ZeroEmpty:         true,
	})
}
