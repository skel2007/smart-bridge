package yandex

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/skel2007/smart-bridge/internal/devices"
)

const (
	headerAuthorization = "Authorization"
	headerRequestID     = "X-Request-Id"
)

type Handler struct {
	gateway     devices.DeviceGateway
	userID      string
	bearerToken string
	mux         *http.ServeMux
}

func NewHandler(gateway devices.DeviceGateway, userID string, bearerToken string) *Handler {
	handler := &Handler{
		gateway:     gateway,
		userID:      userID,
		bearerToken: bearerToken,
	}
	handler.mux = http.NewServeMux()
	handler.mux.HandleFunc("HEAD /v1.0/{$}", handler.serveRoot)
	handler.mux.HandleFunc("POST /v1.0/user/unlink", handler.serveUnlink)
	handler.mux.HandleFunc("GET /v1.0/user/devices", handler.serveDevices)
	handler.mux.HandleFunc("POST /v1.0/user/devices/query", handler.serveDevicesQuery)

	return handler
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !handler.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := r.Header.Get(headerRequestID)
	if requestID == "" {
		http.Error(w, "missing X-Request-Id", http.StatusBadRequest)
		return
	}

	handler.mux.ServeHTTP(w, r)
}

func (handler *Handler) authorized(r *http.Request) bool {
	if handler.bearerToken == "" {
		return false
	}

	return r.Header.Get(headerAuthorization) == "Bearer "+handler.bearerToken
}

func (handler *Handler) serveRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) serveUnlink(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, RequestIDResponse{RequestID: r.Header.Get(headerRequestID)})
}

func (handler *Handler) serveDevices(w http.ResponseWriter, r *http.Request) {
	deviceList, err := handler.gateway.ListDevices(r.Context())
	if err != nil {
		http.Error(w, "list devices failed", http.StatusInternalServerError)
		return
	}

	descriptions := make([]DeviceDescription, 0, len(deviceList))
	for _, device := range deviceList {
		capabilities, err := handler.gateway.ListCapabilities(r.Context(), device.ID)
		if err != nil {
			http.Error(w, "list capabilities failed", http.StatusInternalServerError)
			return
		}

		descriptions = append(descriptions, MapDeviceDescription(device, capabilities))
	}

	writeJSON(w, http.StatusOK, DevicesResponse{
		RequestID: r.Header.Get(headerRequestID),
		Payload: DevicesPayload{
			UserID:  handler.userID,
			Devices: descriptions,
		},
	})
}

func (handler *Handler) serveDevicesQuery(w http.ResponseWriter, r *http.Request) {
	var request DevicesQueryRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	deviceList, err := handler.gateway.ListDevices(r.Context())
	if err != nil {
		http.Error(w, "list devices failed", http.StatusInternalServerError)
		return
	}
	knownDevices := mapDevicesByID(deviceList)

	states := make([]DeviceState, 0, len(request.Devices))
	for _, requestedDevice := range request.Devices {
		if _, ok := knownDevices[requestedDevice.ID]; !ok {
			states = append(states, deviceStateError(requestedDevice.ID, errorCodeDeviceNotFound, "device not found"))
			continue
		}

		capabilities, err := handler.gateway.ListCapabilities(r.Context(), requestedDevice.ID)
		if err != nil {
			states = append(states, deviceStateError(requestedDevice.ID, errorCodeDeviceUnreachable, "device is unreachable"))
			continue
		}

		states = append(states, MapDeviceState(requestedDevice.ID, capabilities))
	}

	writeJSON(w, http.StatusOK, DevicesQueryResponse{
		RequestID: r.Header.Get(headerRequestID),
		Payload: DevicesQueryPayload{
			Devices: states,
		},
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(out); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}

	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func mapDevicesByID(deviceList []devices.Device) map[string]devices.Device {
	result := make(map[string]devices.Device, len(deviceList))
	for _, device := range deviceList {
		result[device.ID] = device
	}

	return result
}

func deviceStateError(deviceID string, code string, message string) DeviceState {
	return DeviceState{
		ID:           deviceID,
		ErrorCode:    code,
		ErrorMessage: message,
	}
}
