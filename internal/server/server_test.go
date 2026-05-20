package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skel2007/smart-bridge/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewMuxRoutesProtocolUnderConfiguredPrefix(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/api/yandex",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1.0/user/devices", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	mux := newMux(cfg, yandexHandlers{
		oauth:    unexpectedHandler(t, "oauth"),
		protocol: handler,
	})
	request := httptest.NewRequest(http.MethodGet, "/api/yandex/v1.0/user/devices", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestNewMuxRoutesProtocolAtRoot(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/",
		},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1.0/user/devices", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	})
	mux := newMux(cfg, yandexHandlers{
		oauth:    unexpectedHandler(t, "oauth"),
		protocol: handler,
	})
	request := httptest.NewRequest(http.MethodGet, "/v1.0/user/devices", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestNewMuxRejectsProtocolOutsideConfiguredPrefix(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/api/yandex",
		},
	}
	mux := newMux(cfg, yandexHandlers{
		oauth:    unexpectedHandler(t, "oauth"),
		protocol: unexpectedHandler(t, "protocol"),
	})
	request := httptest.NewRequest(http.MethodGet, "/v1.0/user/devices", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestNewMuxExposesHealthWhenYandexHandlerIsMountedAtRoot(t *testing.T) {
	cfg := config.Config{
		Yandex: config.YandexConfig{
			PathPrefix: "/",
		},
	}
	mux := newMux(cfg, yandexHandlers{
		oauth:    unexpectedHandler(t, "oauth"),
		protocol: unexpectedHandler(t, "protocol"),
	})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
}

func TestNewMuxRoutesOAuthBeforeProtocol(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		requestURL string
	}{
		{
			name:       "under configured prefix",
			prefix:     "/api/yandex",
			requestURL: "/api/yandex/oauth/authorize",
		},
		{
			name:       "at root",
			prefix:     "/",
			requestURL: "/oauth/authorize",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				Yandex: config.YandexConfig{
					PathPrefix: tt.prefix,
				},
			}
			oauthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/authorize", r.URL.Path)
				w.WriteHeader(http.StatusAccepted)
			})
			protocolHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusConflict)
			})
			mux := newMux(cfg, yandexHandlers{
				oauth:    oauthHandler,
				protocol: protocolHandler,
			})
			request := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, request)

			require.Equal(t, http.StatusAccepted, response.Code)
		})
	}
}

func unexpectedHandler(t *testing.T, name string) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		t.Helper()
		require.FailNowf(t, "unexpected handler called", "%s received %s", name, r.URL.Path)
	})
}
