package yandex

import (
	"context"
	"encoding/json"
	"errors"
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

func TestHandlerDevices(t *testing.T) {
	gateway := &fakeGateway{
		devices: []devices.Device{
			{ID: "light-1", Name: "Desk light", Type: devices.DeviceTypeLight, Online: true},
		},
		capabilities: map[string][]devices.Capability{
			"light-1": {
				devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
			},
		},
	}
	handler := newTestHandler(gateway)
	request := newHandlerRequest(http.MethodGet, "/v1.0/user/devices", "")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
		"request_id": "request-1",
		"payload": {
			"user_id": "bridge-user",
			"devices": [
				{
					"id": "light-1",
					"name": "Desk light",
					"status_info": {
						"reportable": false
					},
					"type": "devices.types.light",
					"capabilities": [
						{
							"type": "devices.capabilities.on_off",
							"retrievable": true,
							"reportable": false
						}
					]
				}
			]
		}
	}`, response.Body.String())
	require.Equal(t, []string{"light-1"}, gateway.listCapabilityCalls)
}

func TestHandlerDevicesReturnsRequestLevelErrors(t *testing.T) {
	tests := []struct {
		name       string
		gateway    *fakeGateway
		wantStatus int
	}{
		{
			name: "list devices fails",
			gateway: &fakeGateway{
				listDevicesErr: errors.New("upstream unavailable"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "list capabilities fails",
			gateway: &fakeGateway{
				devices: []devices.Device{
					{ID: "light-1", Name: "Desk light", Type: devices.DeviceTypeLight},
				},
				capabilityErrors: map[string]error{
					"light-1": errors.New("upstream unavailable"),
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(tt.gateway)
			request := newHandlerRequest(http.MethodGet, "/v1.0/user/devices", "")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, tt.wantStatus, response.Code)
		})
	}
}

func TestHandlerDevicesQueryReturnsPerDeviceErrors(t *testing.T) {
	gateway := &fakeGateway{
		devices: []devices.Device{
			{ID: "light-1", Name: "Desk light", Type: devices.DeviceTypeLight},
			{ID: "light-2", Name: "Floor light", Type: devices.DeviceTypeLight},
		},
		capabilities: map[string][]devices.Capability{
			"light-1": {
				devices.NewOnOffCapability(devices.CapabilityInstancePower, true),
			},
		},
		capabilityErrors: map[string]error{
			"light-2": errors.New("upstream unavailable"),
		},
	}
	handler := newTestHandler(gateway)
	request := newHandlerRequest(http.MethodPost, "/v1.0/user/devices/query", `{
		"devices": [
			{"id": "light-1"},
			{"id": "missing-light"},
			{"id": "light-2"}
		]
	}`)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
		"request_id": "request-1",
		"payload": {
			"devices": [
				{
					"id": "light-1",
					"capabilities": [
						{
							"type": "devices.capabilities.on_off",
							"state": {
								"instance": "on",
								"value": true
							}
						}
					]
				},
				{
					"id": "missing-light",
					"error_code": "DEVICE_NOT_FOUND",
					"error_message": "device not found"
				},
				{
					"id": "light-2",
					"error_code": "DEVICE_UNREACHABLE",
					"error_message": "device is unreachable"
				}
			]
		}
	}`, response.Body.String())
}

func TestHandlerDevicesQueryReturnsRequestLevelErrors(t *testing.T) {
	tests := []struct {
		name       string
		gateway    *fakeGateway
		body       string
		wantStatus int
	}{
		{
			name:       "malformed json",
			gateway:    &fakeGateway{},
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "trailing json token",
			gateway:    &fakeGateway{},
			body:       `{"devices": []}{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "list devices fails",
			gateway: &fakeGateway{
				listDevicesErr: errors.New("upstream unavailable"),
			},
			body:       `{"devices": []}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(tt.gateway)
			request := newHandlerRequest(http.MethodPost, "/v1.0/user/devices/query", tt.body)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, tt.wantStatus, response.Code)
		})
	}
}

func TestHandlerDevicesActionReturnsPerDeviceResults(t *testing.T) {
	gateway := &fakeGateway{
		devices: []devices.Device{
			{ID: "light-1", Name: "Desk light", Type: devices.DeviceTypeLight},
			{ID: "light-invalid", Name: "Invalid light", Type: devices.DeviceTypeLight},
			{ID: "light-unreachable", Name: "Unreachable light", Type: devices.DeviceTypeLight},
		},
		sendErrors: map[string]error{
			"light-unreachable": errors.New("upstream unavailable"),
		},
	}
	body := mustJSON(t, DevicesActionRequest{
		Payload: DevicesActionPayload{
			Devices: []DeviceAction{
				{
					ID: "light-1",
					Capabilities: []CapabilityAction{
						newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, true),
					},
				},
				{
					ID: "missing-light",
					Capabilities: []CapabilityAction{
						newTestCapabilityAction(capabilityTypeOnOff, capabilityInstanceOn, true),
					},
				},
				{
					ID: "light-invalid",
					Capabilities: []CapabilityAction{
						newTestCapabilityAction(capabilityTypeRange, capabilityInstanceBrightness, 101),
					},
				},
				{
					ID: "light-unreachable",
					Capabilities: []CapabilityAction{
						newTestCapabilityAction(capabilityTypeRange, capabilityInstanceBrightness, 42),
					},
				},
			},
		},
	})
	handler := newTestHandler(gateway)
	request := newHandlerRequest(http.MethodPost, "/v1.0/user/devices/action", body)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)

	var actionResponse DevicesActionResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &actionResponse))
	require.Equal(t, "request-1", actionResponse.RequestID)
	require.Len(t, actionResponse.Payload.Devices, 4)
	require.Equal(t, actionStatusDone, actionResponse.Payload.Devices[0].Capabilities[0].State.ActionResult.Status)
	require.Equal(t, errorCodeDeviceNotFound, actionResponse.Payload.Devices[1].ActionResult.ErrorCode)
	require.Equal(t, errorCodeInvalidValue, actionResponse.Payload.Devices[2].Capabilities[0].State.ActionResult.ErrorCode)
	require.Equal(t, errorCodeDeviceUnreachable, actionResponse.Payload.Devices[3].Capabilities[0].State.ActionResult.ErrorCode)
	require.Equal(t, []devices.CapabilityCommand{
		devices.NewOnOffCommand(devices.CapabilityInstancePower, true),
	}, gateway.sentCommands["light-1"])
	require.Equal(t, []devices.CapabilityCommand{
		devices.NewRangeCommand(devices.CapabilityInstanceBrightness, 42),
	}, gateway.sentCommands["light-unreachable"])
	require.NotContains(t, gateway.sentCommands, "light-invalid")
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

func mustJSON(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return string(data)
}

type fakeGateway struct {
	devices             []devices.Device
	listDevicesErr      error
	capabilities        map[string][]devices.Capability
	capabilityErrors    map[string]error
	listCapabilityCalls []string
	sendErrors          map[string]error
	sentCommands        map[string][]devices.CapabilityCommand
}

func (gateway *fakeGateway) ListDevices(ctx context.Context) ([]devices.Device, error) {
	return gateway.devices, gateway.listDevicesErr
}

func (gateway *fakeGateway) ListCapabilities(ctx context.Context, deviceID string) ([]devices.Capability, error) {
	gateway.listCapabilityCalls = append(gateway.listCapabilityCalls, deviceID)
	if err := gateway.capabilityErrors[deviceID]; err != nil {
		return nil, err
	}

	return gateway.capabilities[deviceID], nil
}

func (gateway *fakeGateway) SendCommands(ctx context.Context, deviceID string, commands []devices.CapabilityCommand) error {
	if gateway.sentCommands == nil {
		gateway.sentCommands = make(map[string][]devices.CapabilityCommand)
	}
	gateway.sentCommands[deviceID] = append([]devices.CapabilityCommand(nil), commands...)
	if err := gateway.sendErrors[deviceID]; err != nil {
		return err
	}

	return nil
}
