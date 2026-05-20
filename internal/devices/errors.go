package devices

import (
	"errors"
	"fmt"
)

var ErrCapabilityNotSupported = errors.New("capability not supported")

func NewCapabilityNotSupportedError(format string, args ...any) error {
	return capabilityNotSupportedError{Message: fmt.Sprintf(format, args...)}
}

type capabilityNotSupportedError struct {
	Message string
}

func (err capabilityNotSupportedError) Error() string {
	return err.Message
}

func (err capabilityNotSupportedError) Unwrap() error {
	return ErrCapabilityNotSupported
}
