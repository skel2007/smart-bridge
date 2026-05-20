package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skel2007/smart-bridge/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewMuxMountsYandexHandlerUnderConfiguredPrefix(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/api/yandex",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1.0/user/devices", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	mux := newMux(cfg, handler)
	request := httptest.NewRequest(http.MethodGet, "/api/yandex/v1.0/user/devices", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestNewMuxCanMountYandexHandlerAtRoot(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1.0/user/devices", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	mux := newMux(cfg, handler)
	request := httptest.NewRequest(http.MethodGet, "/v1.0/user/devices", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestNewMuxDoesNotExposeYandexHandlerOutsideConfiguredPrefix(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/api/yandex",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux := newMux(cfg, handler)
	request := httptest.NewRequest(http.MethodGet, "/v1.0/user/devices", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestNewMuxExposesHealthOutsideYandexPrefix(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/api/yandex",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux := newMux(cfg, handler)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestNewMuxExposesHealthWhenYandexHandlerIsMountedAtRoot(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux := newMux(cfg, handler)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}
