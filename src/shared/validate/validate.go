// Package validate: tag driven validation for request structs
// wraps go-playground/validator, nonblank rejects blank strings,
// Optional fields validate only when set
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

// Builds validator with platform conventions
//
// # Returns:
//
//   - Ready validator
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

// Unwraps Optional field for tag validation
//
// # Inputs:
//
//   - field [reflect.Value] the Optional field
//
// # Returns:
//
//   - Carried value or nil when absent or null
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

// Validates correctness of struct reporting first failure
//
// # Inputs:
//
//   - s [any] the struct to be validated
//
// # Returns:
//
//   - First failure reported by validator or nil
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

// Maps field error to readable failure reason
//
// # Inputs:
//
//   - fe [validator.FieldError] the failed field
//
// # Returns:
//
//   - Readable reason
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
