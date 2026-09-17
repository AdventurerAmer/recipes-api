package errs

type UserError struct {
	TraceId string `json:"traceId"`
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Fields  Fields `json:"fields,omitempty"`
}

func NewUserError(traceId string, err *Error) UserError {
	return UserError{
		TraceId: traceId,
		Code:    err.Code,
		Message: err.Message,
		Fields:  err.Fields,
	}
}
