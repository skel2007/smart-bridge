package yandex

const (
	errorCodeInvalidValue              = "INVALID_VALUE"
	errorCodeNotSupportedInCurrentMode = "NOT_SUPPORTED_IN_CURRENT_MODE"
	errorCodeDeviceUnreachable         = "DEVICE_UNREACHABLE"
	errorCodeDeviceNotFound            = "DEVICE_NOT_FOUND"
)

type ActionMappingError struct {
	Code    string
	Message string
	Cause   error
}

func (err ActionMappingError) Error() string {
	return err.Message
}

func (err ActionMappingError) Unwrap() error {
	return err.Cause
}
