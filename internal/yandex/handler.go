package yandex

import (
	"encoding/json"
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
	handler.mux.HandleFunc("HEAD /v1.0/", handler.serveRoot)
	handler.mux.HandleFunc("POST /v1.0/user/unlink", handler.serveUnlink)

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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
