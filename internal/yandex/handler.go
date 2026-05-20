package yandex

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/skel2007/smart-bridge/internal/devices"
)

const (
	headerAuthorization = "Authorization"
	headerRequestID     = "X-Request-Id"

	// maxRequestBodyBytes limits Yandex JSON requests to 1 MiB.
	maxRequestBodyBytes = 1 << 20
)

type Handler struct {
	gateway     devices.DeviceGateway
	userID      string
	bearerToken string
	logger      *slog.Logger
	mux         *http.ServeMux
}

type HandlerConfig struct {
	UserID      string
	BearerToken string
}

type Option func(*Handler)

func WithLogger(logger *slog.Logger) Option {
	return func(handler *Handler) {
		if logger != nil {
			handler.logger = logger
		}
	}
}

func NewHandler(gateway devices.DeviceGateway, cfg HandlerConfig, options ...Option) *Handler {
	handler := &Handler{
		gateway:     gateway,
		userID:      cfg.UserID,
		bearerToken: cfg.BearerToken,
		logger:      slog.New(slog.DiscardHandler),
		mux:         http.NewServeMux(),
	}
	for _, option := range options {
		option(handler)
	}

	handler.mux.HandleFunc("HEAD /v1.0/{$}", handler.serveRoot)
	handler.mux.HandleFunc("POST /v1.0/user/unlink", handler.serveUnlink)
	handler.mux.HandleFunc("GET /v1.0/user/devices", handler.serveDevices)
	handler.mux.HandleFunc("POST /v1.0/user/devices/query", handler.serveDevicesQuery)
	handler.mux.HandleFunc("POST /v1.0/user/devices/action", handler.serveDevicesAction)

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

func (handler *Handler) serveRoot(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) serveUnlink(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, RequestIDResponse{RequestID: r.Header.Get(headerRequestID)})
}

func (handler *Handler) serveDevices(w http.ResponseWriter, r *http.Request) {
	deviceList, err := handler.gateway.ListDevices(r.Context())
	if err != nil {
		handler.logRequest(r, slog.LevelWarn, "list yandex devices failed", "error", err)
		http.Error(w, "list devices failed", http.StatusInternalServerError)
		return
	}

	descriptions := make([]DeviceDescription, 0, len(deviceList))
	for _, device := range deviceList {
		capabilities, err := handler.gateway.ListCapabilities(r.Context(), device.ID)
		if err != nil {
			handler.logRequest(r, slog.LevelWarn, "load yandex device capabilities failed",
				"device_id", device.ID,
				"error", err,
			)
			continue
		}

		descriptions = append(descriptions, mapDeviceDescription(device, capabilities))
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
		handler.logRequest(r, slog.LevelWarn, "query yandex devices failed", "error", err)
		writeJSON(w, http.StatusOK, DevicesQueryResponse{
			RequestID: r.Header.Get(headerRequestID),
			Payload: DevicesQueryPayload{
				Devices: newDeviceStateErrors(request.Devices, errorCodeDeviceUnreachable, "device is unreachable"),
			},
		})
		return
	}
	knownDeviceIDs := mapDeviceIDs(deviceList)

	states := make([]DeviceState, 0, len(request.Devices))
	for _, requestedDevice := range request.Devices {
		if _, ok := knownDeviceIDs[requestedDevice.ID]; !ok {
			handler.logDeviceError(r, requestedDevice.ID, errorCodeDeviceNotFound, "query requested unknown device", nil)
			states = append(states, newDeviceStateError(requestedDevice.ID, errorCodeDeviceNotFound, "device not found"))
			continue
		}

		capabilities, err := handler.gateway.ListCapabilities(r.Context(), requestedDevice.ID)
		if err != nil {
			handler.logDeviceError(r, requestedDevice.ID, errorCodeDeviceUnreachable, "query yandex device capabilities failed", err)
			states = append(states, newDeviceStateError(requestedDevice.ID, errorCodeDeviceUnreachable, "device is unreachable"))
			continue
		}

		states = append(states, mapDeviceState(requestedDevice.ID, capabilities))
	}

	writeJSON(w, http.StatusOK, DevicesQueryResponse{
		RequestID: r.Header.Get(headerRequestID),
		Payload: DevicesQueryPayload{
			Devices: states,
		},
	})
}

func (handler *Handler) serveDevicesAction(w http.ResponseWriter, r *http.Request) {
	var request DevicesActionRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	deviceList, err := handler.gateway.ListDevices(r.Context())
	if err != nil {
		handler.logRequest(r, slog.LevelWarn, "prepare yandex device action failed", "error", err)
		writeJSON(w, http.StatusOK, DevicesActionResponse{
			RequestID: r.Header.Get(headerRequestID),
			Payload: DevicesActionResults{
				Devices: newDeviceActionErrors(request.Payload.Devices, errorCodeDeviceUnreachable, "device is unreachable"),
			},
		})
		return
	}
	knownDeviceIDs := mapDeviceIDs(deviceList)

	results := make([]DeviceActionResult, 0, len(request.Payload.Devices))
	for _, deviceAction := range request.Payload.Devices {
		if _, ok := knownDeviceIDs[deviceAction.ID]; !ok {
			handler.logDeviceError(r, deviceAction.ID, errorCodeDeviceNotFound, "action requested unknown device", nil)
			results = append(results, newDeviceActionError(deviceAction.ID, errorCodeDeviceNotFound, "device not found"))
			continue
		}

		commands, err := mapDeviceActionCommands(deviceAction)
		if err != nil {
			if mappingErr, ok := errors.AsType[actionMappingError](err); ok {
				handler.logDeviceError(r, deviceAction.ID, mappingErr.Code, "map yandex device action failed", err)
				results = append(results, newDeviceCapabilityResults(deviceAction, newActionError(mappingErr.Code, mappingErr.Message)))
				continue
			}

			handler.logDeviceError(r, deviceAction.ID, errorCodeInvalidValue, "map yandex device action failed", err)
			results = append(results, newDeviceCapabilityResults(deviceAction, newActionError(errorCodeInvalidValue, err.Error())))
			continue
		}

		if err := handler.gateway.SendCommands(r.Context(), deviceAction.ID, commands); err != nil {
			errorCode, errorMessage := mapActionSendError(err)
			handler.logDeviceError(r, deviceAction.ID, errorCode, "send yandex device action failed", err)
			results = append(results, newDeviceCapabilityResults(deviceAction, newActionError(errorCode, errorMessage)))
			continue
		}

		results = append(results, newDeviceCapabilityResults(deviceAction, ActionResult{Status: actionStatusDone}))
	}

	writeJSON(w, http.StatusOK, DevicesActionResponse{
		RequestID: r.Header.Get(headerRequestID),
		Payload: DevicesActionResults{
			Devices: results,
		},
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(out); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return false
		}

		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return false
		}

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

func mapDeviceIDs(deviceList []devices.Device) map[string]struct{} {
	result := make(map[string]struct{}, len(deviceList))
	for _, device := range deviceList {
		result[device.ID] = struct{}{}
	}

	return result
}

func (handler *Handler) logDeviceError(r *http.Request, deviceID string, code string, message string, err error) {
	level := slog.LevelDebug
	if code == errorCodeDeviceUnreachable {
		level = slog.LevelWarn
	}

	attrs := []any{
		"device_id", deviceID,
		"error_code", code,
	}
	if err != nil {
		attrs = append(attrs, "error", err)
	}

	handler.logRequest(r, level, message, attrs...)
}

func (handler *Handler) logRequest(r *http.Request, level slog.Level, message string, attrs ...any) {
	attrs = append([]any{
		"request_id", r.Header.Get(headerRequestID),
		"method", r.Method,
		"path", r.URL.Path,
	}, attrs...)

	handler.logger.Log(r.Context(), level, message, attrs...)
}
