package cloud

import (
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (err *APIError) Error() string {
	parts := []string{"tuya api error"}
	if err.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("status=%d", err.StatusCode))
	}
	if err.Code != "" {
		parts = append(parts, "code="+err.Code)
	}
	if err.Message != "" {
		parts = append(parts, "message="+err.Message)
	}

	return strings.Join(parts, ": ")
}
