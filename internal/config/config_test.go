package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Config
	}{
		{
			name: "loads platform sections",
			content: `
http:
  port: 8080
tuya:
  endpoint: https://example.com
  client_id: client-id
  client_secret: client-secret
yandex:
  user_id: bridge-user
  bearer_token: bearer-token
  path_prefix: /custom/yandex
  oauth:
    client_id: oauth-client
    client_secret: oauth-secret
`,
			want: Config{
				HTTP: HTTPConfig{
					Port: 8080,
				},
				Tuya: TuyaConfig{
					Endpoint:     "https://example.com",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
				Yandex: YandexConfig{
					UserID:      "bridge-user",
					BearerToken: "bearer-token",
					PathPrefix:  "/custom/yandex",
					OAuth: OAuthConfig{
						ClientID:     "oauth-client",
						ClientSecret: "oauth-secret",
					},
				},
			},
		},
		{
			name: "applies defaults",
			content: `
tuya:
  client_id: client-id
  client_secret: client-secret
`,
			want: Config{
				Tuya: TuyaConfig{
					Endpoint:     DefaultTuyaEndpoint,
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
				Yandex: YandexConfig{
					PathPrefix: DefaultYandexPathPrefix,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.content)

			cfg, err := Load(path)

			require.NoError(t, err)
			require.Equal(t, tt.want, cfg)
		})
	}
}

func TestLoadReturnsClearErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    func(t *testing.T) string
		wantErr string
	}{
		{
			name: "invalid YAML",
			path: func(t *testing.T) string {
				t.Helper()

				return writeConfig(t, `tuya: [`)
			},
			wantErr: "invalid YAML in config file",
		},
		{
			name: "missing file",
			path: func(t *testing.T) string {
				t.Helper()

				return filepath.Join(t.TempDir(), "missing.yaml")
			},
			wantErr: "config file not found:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(tt.path(t))

			require.Error(t, err)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestHTTPConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     HTTPConfig
		wantErr string
	}{
		{
			name: "valid",
			cfg: HTTPConfig{
				Port: 8080,
			},
		},
		{
			name: "default ephemeral port",
			cfg:  HTTPConfig{},
		},
		{
			name: "negative port",
			cfg: HTTPConfig{
				Port: -1,
			},
			wantErr: "http.port must be between 0 and 65535",
		},
		{
			name: "port above maximum",
			cfg: HTTPConfig{
				Port: 65536,
			},
			wantErr: "http.port must be between 0 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestTuyaConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     TuyaConfig
		wantErr string
	}{
		{
			name: "valid",
			cfg: TuyaConfig{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			},
		},
		{
			name: "missing client ID",
			cfg: TuyaConfig{
				ClientSecret: "super-secret",
			},
			wantErr: "tuya.client_id is required",
		},
		{
			name: "missing client secret",
			cfg: TuyaConfig{
				ClientID: "client-id",
			},
			wantErr: "tuya.client_secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tt.wantErr)
			require.NotContains(t, err.Error(), "super-secret")
		})
	}
}

func TestYandexConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*YandexConfig)
		wantErr string
	}{
		{
			name: "valid",
		},
		{
			name: "missing user ID",
			mutate: func(cfg *YandexConfig) {
				cfg.UserID = ""
				cfg.BearerToken = "super-secret"
			},
			wantErr: "yandex.user_id is required",
		},
		{
			name: "missing bearer token",
			mutate: func(cfg *YandexConfig) {
				cfg.BearerToken = ""
			},
			wantErr: "yandex.bearer_token is required",
		},
		{
			name: "missing path prefix",
			mutate: func(cfg *YandexConfig) {
				cfg.BearerToken = "super-secret"
				cfg.PathPrefix = ""
			},
			wantErr: "yandex.path_prefix is required",
		},
		{
			name: "path prefix without leading slash",
			mutate: func(cfg *YandexConfig) {
				cfg.BearerToken = "super-secret"
				cfg.PathPrefix = "api/yandex"
			},
			wantErr: "yandex.path_prefix must start with /",
		},
		{
			name: "path prefix with trailing slash",
			mutate: func(cfg *YandexConfig) {
				cfg.BearerToken = "super-secret"
				cfg.PathPrefix = "/api/yandex/"
			},
			wantErr: "yandex.path_prefix must not end with /",
		},
		{
			name: "missing OAuth client ID",
			mutate: func(cfg *YandexConfig) {
				cfg.BearerToken = "super-secret"
				cfg.OAuth.ClientID = ""
			},
			wantErr: "yandex.oauth.client_id is required",
		},
		{
			name: "missing OAuth client secret",
			mutate: func(cfg *YandexConfig) {
				cfg.OAuth.ClientSecret = ""
			},
			wantErr: "yandex.oauth.client_secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validYandexConfig()
			if tt.mutate != nil {
				tt.mutate(&cfg)
			}

			err := cfg.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tt.wantErr)
			require.NotContains(t, err.Error(), "super-secret")
		})
	}
}

func TestOAuthConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     OAuthConfig
		wantErr string
	}{
		{
			name: "valid",
			cfg: OAuthConfig{
				ClientID:     "oauth-client",
				ClientSecret: "oauth-secret",
			},
		},
		{
			name: "missing client ID",
			cfg: OAuthConfig{
				ClientSecret: "super-secret",
			},
			wantErr: "yandex.oauth.client_id is required",
		},
		{
			name: "missing client secret",
			cfg: OAuthConfig{
				ClientID: "oauth-client",
			},
			wantErr: "yandex.oauth.client_secret is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tt.wantErr)
			require.NotContains(t, err.Error(), "super-secret")
		})
	}
}

func validYandexConfig() YandexConfig {
	return YandexConfig{
		UserID:      "bridge-user",
		BearerToken: "bearer-token",
		PathPrefix:  DefaultYandexPathPrefix,
		OAuth: OAuthConfig{
			ClientID:     "oauth-client",
			ClientSecret: "oauth-secret",
		},
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
