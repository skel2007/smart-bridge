package yandex

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActionMappingErrorUnwrap(t *testing.T) {
	cause := errors.New("decode failed")
	err := ActionMappingError{
		Code:    errorCodeInvalidValue,
		Message: "invalid action value",
		Cause:   cause,
	}

	require.EqualError(t, err, "invalid action value")
	require.ErrorIs(t, err, cause)
}
