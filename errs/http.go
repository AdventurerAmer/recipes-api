package errs

import (
	"net/http"
)

func HTTPStatus(code Code) int {
	switch code {
	case CodeInternal:
		return http.StatusInternalServerError
	case CodeValidation:
		return http.StatusBadRequest
	case CodeResourceNotFound:
		return http.StatusNotFound
	case CodeResourceAlreadyExists:
		return http.StatusConflict
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodeUnsupportedFormat:
		return http.StatusUnsupportedMediaType
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}
