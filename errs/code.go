package errs

type Code string

const (
	CodeInternal              Code = "INTERNAL"
	CodeValidation            Code = "VALIDATION"
	CodeResourceNotFound      Code = "RESOURCE_NOT_FOUND"
	CodeResourceAlreadyExists Code = "RESOURCE_ALREADY_EXISTS"
	CodeTimeout               Code = "TIMEOUT"
	CodeUnsupportedFormat     Code = "UNSUPPORTED_FORMAT"
	CodeServiceUnavailable    Code = "SERVICE_UNAVAILABLE"
)
