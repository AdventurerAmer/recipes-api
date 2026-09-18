package errs

import "errors"

func NewInternal(err error) *Error {
	return Wrap(err, CodeInternal, "internal server error")
}

type Fields = map[string]string

func NewValidation(fields Fields) *Error {
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

func NewNetwork(err error) *Error {
	return Wrap(err, CodeNetwork, "network failure")
}

func NewConflict(err error, message string) *Error {
	return Wrap(err, CodeConflict, message)
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

func IsNetwork(err error) bool {
	return Is(err, CodeNetwork)
}

func IsConflict(err error) bool {
	return Is(err, CodeConflict)
}
