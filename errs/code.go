package errs

type Code string

const (
	CodeInternal              Code = "INTERNAL"
	CodeValidation            Code = "VALIDATION"
	CodeResourceNotFound      Code = "RESOURCE_NOT_FOUND"
	CodeResourceAlreadyExists Code = "RESOURCE_ALREADY_EXISTS"
	CodeTimeout               Code = "TIMEOUT"
	CodeNetwork               Code = "NETWORK"
	CodeUnsupportedFormat     Code = "UNSUPPORTED_FORMAT"
	CodeServiceUnavailable    Code = "SERVICE_UNAVAILABLE"
	CodeConflict              Code = "CONFLICT"
	CodeAuthentication        Code = "AUTHENTICATION"
	CodeAuthorization         Code = "AUTHORIZATION"
)

func IsRetryCode(code Code) bool {
	switch code {
	case CodeInternal,
		CodeNetwork,
		CodeTimeout,
		CodeServiceUnavailable,
		CodeConflict:
		return true
	}
	return false
}
