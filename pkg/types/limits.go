package types

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// MaxLength returns the max=N limit in the validate tag of a field of the model. The forms read their
// maxlength from here, so what the browser allows is exactly what the server accepts.
func MaxLength(model any, field string) (int, error) {
	structField, ok := reflect.TypeOf(model).FieldByName(field)
	if !ok {
		return 0, fmt.Errorf("%T has no field %s", model, field)
	}

	for _, rule := range strings.Split(structField.Tag.Get("validate"), ",") {
		if value, found := strings.CutPrefix(rule, "max="); found {
			return strconv.Atoi(value)
		}
	}
	return 0, fmt.Errorf("%T.%s has no max in its validate tag", model, field)
}

// ActivityMax returns the longest allowed value of an activity field, or the most participants.
func ActivityMax(field string) (int, error) {
	return MaxLength(Activity{}, field)
}

// CertificationMax returns the longest allowed value of a certification field, or the most participants.
func CertificationMax(field string) (int, error) {
	return MaxLength(Certification{}, field)
}
