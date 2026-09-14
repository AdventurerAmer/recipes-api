package errs

import "errors"

func NewInternal(err error) *Error {
	return Wrap(err, CodeInternal, "internal server error")
}

func NewValidation(fields map[string]string) *Error {
	err := New(CodeValidation, "one or more invalid fields")
	err.Fields = fields
	return err
}

func NewNotFound(err error, message string) *Error {
	return Wrap(err, CodeResourceNotFound, message)
}

func NewAlreadyExists(err error, message string) *Error {
	return Wrap(err, CodeResourceAlreadyExists, message)
}

func NewTimeout(err error) *Error {
	return Wrap(err, CodeTimeout, "timeout")
}

func NewUnsupportedFormat(err error) *Error {
	return Wrap(err, CodeUnsupportedFormat, "unsupported format")
}

func NewServiceUnavailable(err error) *Error {
	return Wrap(err, CodeServiceUnavailable, "service unavailable")
}

func Is(err error, code Code) bool {
	var e *Error
	if errors.As(err, &e) && e.Code == code {
		return true
	}
	return false
}

func IsInternal(err error) bool {
	return Is(err, CodeInternal)
}

func IsNotFound(err error) bool {
	return Is(err, CodeResourceNotFound)
}

func IsValidation(err error) bool {
	return Is(err, CodeValidation)
}

func IsAlreadyExists(err error) bool {
	return Is(err, CodeResourceAlreadyExists)
}

func IsTimeout(err error) bool {
	return Is(err, CodeTimeout)
}

func IsUnsupportedFormat(err error) bool {
	return Is(err, CodeUnsupportedFormat)
}

func IsServiceUnavailable(err error) bool {
	return Is(err, CodeServiceUnavailable)
}
