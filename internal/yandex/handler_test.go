package yandex

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/stretchr/testify/require"
)

func TestHandlerHEADRoot(t *testing.T) {
	handler := newTestHandler(&fakeGateway{})
	request := newHandlerRequest(http.MethodHead, "/v1.0/", "")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Empty(t, response.Body.String())
}

func TestHandlerRequestValidation(t *testing.T) {
	tests := []struct {
		name       string
		configure  func(request *http.Request)
		wantStatus int
	}{
		{
			name: "missing authorization",
			configure: func(request *http.Request) {
				request.Header.Del(headerAuthorization)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid authorization",
			configure: func(request *http.Request) {
				request.Header.Set(headerAuthorization, "Bearer wrong-token")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing request id",
			configure: func(request *http.Request) {
				request.Header.Del(headerRequestID)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown route",
			configure:  func(request *http.Request) {},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "wrong method",
			configure: func(request *http.Request) {
				request.Method = http.MethodGet
				request.URL.Path = "/v1.0/"
			},
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(&fakeGateway{})
			request := newHandlerRequest(http.MethodGet, "/unknown", "")
			tt.configure(request)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, tt.wantStatus, response.Code)
		})
	}
}

func TestHandlerUnlink(t *testing.T) {
	handler := newTestHandler(&fakeGateway{})
	request := newHandlerRequest(http.MethodPost, "/v1.0/user/unlink", "")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"request_id":"request-1"}`, response.Body.String())
}

func newTestHandler(gateway devices.DeviceGateway) *Handler {
	return NewHandler(gateway, "bridge-user", "secret-token")
}

func newHandlerRequest(method string, path string, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set(headerAuthorization, "Bearer secret-token")
	request.Header.Set(headerRequestID, "request-1")

	return request
}

type fakeGateway struct{}

func (gateway *fakeGateway) ListDevices(ctx context.Context) ([]devices.Device, error) {
	return nil, nil
}

func (gateway *fakeGateway) ListCapabilities(ctx context.Context, deviceID string) ([]devices.Capability, error) {
	return nil, nil
}

func (gateway *fakeGateway) SendCommands(ctx context.Context, deviceID string, commands []devices.CapabilityCommand) error {
	return nil
}
