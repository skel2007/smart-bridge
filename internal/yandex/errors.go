package yandex

const (
	errorCodeInvalidValue              = "INVALID_VALUE"
	errorCodeNotSupportedInCurrentMode = "NOT_SUPPORTED_IN_CURRENT_MODE"
	errorCodeDeviceUnreachable         = "DEVICE_UNREACHABLE"
	errorCodeDeviceNotFound            = "DEVICE_NOT_FOUND"
)

type actionMappingError struct {
	Code    string
	Message string
	Cause   error
}

func (err actionMappingError) Error() string {
	return err.Message
}

func (err actionMappingError) Unwrap() error {
	return err.Cause
}
