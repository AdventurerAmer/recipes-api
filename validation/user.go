package validation

import (
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

func StrongPasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	length := utf8.RuneCountInString(password)
	if length < 12 || length > 32 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}
