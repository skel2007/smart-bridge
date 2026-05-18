package yandex

const (
	errorCodeInvalidValue              = "INVALID_VALUE"
	errorCodeNotSupportedInCurrentMode = "NOT_SUPPORTED_IN_CURRENT_MODE"
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
