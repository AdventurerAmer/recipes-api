package validation

import (
	"reflect"
	"strings"

	"github.com/AdventurerAmer/recipes-api/errs"
	"github.com/go-playground/validator/v10"
)

func Validate(s any) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	v.RegisterValidation("strong_password", StrongPasswordValidator)

	if err := v.Struct(s); err != nil {
		fields := formatValidationErrors(err)
		return errs.NewValidation(fields)
	}
	return nil
}

func formatValidationErrors(err error) errs.Fields {
	fields := make(errs.Fields)
	for _, err := range err.(validator.ValidationErrors) {
		key := err.Namespace()
		field := key
		if _, after, ok := strings.Cut(key, "."); ok {
			field = after
		}
		fields[field] = getErrorMessage(err)
	}
	return fields
}

func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + err.Param() + " characters"
	case "max":
		return "Must be at most " + err.Param() + " characters"
	case "oneof":
		return "Must be one of: " + err.Param()
	case "url":
		return "Must be a valid URL"
	case "len":
		return "Must be exactly " + err.Param() + " characters"
	default:
		return err.Error()
	}
}
