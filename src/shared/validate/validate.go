// Package validate runs tag-driven validation for request structs. It wraps
// github.com/go-playground/validator with the platform's conventions: the
// nonblank tag rejects empty or whitespace-only strings, and
// jsonutils.Optional fields validate the value they carry only when one is
// set.
package validate

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	jsonutils "IDIG4110/shared/json-utils"
)

var std = newStd()

func newStd() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		return strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
	})
	_ = v.RegisterValidation("nonblank", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
	v.RegisterCustomTypeFunc(unwrapOptional,
		jsonutils.Optional[string]{},
		jsonutils.Optional[int]{},
		jsonutils.Optional[bool]{},
		jsonutils.Optional[any]{},
		jsonutils.Optional[map[string]any]{},
		jsonutils.Optional[time.Time]{},
	)
	return v
}

// unwrapOptional returns the value an Optional field carries, or nil when the
// field is absent or explicitly null so the remaining tags skip it.
func unwrapOptional(field reflect.Value) any {
	set := field.FieldByName("Set")
	val := field.FieldByName("Value")
	if !set.IsValid() || set.Kind() != reflect.Bool || !val.IsValid() || val.Kind() != reflect.Ptr {
		return field.Interface()
	}
	if !set.Bool() || val.IsNil() {
		return nil
	}
	return val.Elem().Interface()
}

// Struct validates s against its validate tags and reports the first
// failure, e.g. "home_id is not a uuid".
func Struct(s any) error {
	err := std.Struct(s)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) || len(verrs) == 0 {
		return err
	}
	fe := verrs[0]
	return fmt.Errorf("%s %s", fe.Field(), reason(fe))
}

func reason(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "uuid":
		return "is not a uuid"
	case "nonblank":
		return "cannot be blank"
	case "max":
		return "is too long"
	case "oneof":
		return "must be one of " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default:
		return "failed " + fe.Tag() + " validation"
	}
}
