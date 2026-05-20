package server

import (
	"testing"

	"github.com/skel2007/smart-bridge/internal/config"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigRequiresHTTPAndPlatformSections(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr string
	}{
		{
			name: "valid",
			cfg: config.Config{
				HTTP: config.HTTPConfig{
					Port: 0,
				},
				Tuya: config.TuyaConfig{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
				Yandex: config.YandexConfig{
					UserID:      "bridge-user",
					BearerToken: "bearer-token",
					PathPrefix:  config.DefaultYandexPathPrefix,
					OAuth: config.OAuthConfig{
						ClientID:     "oauth-client",
						ClientSecret: "oauth-secret",
					},
				},
			},
		},
		{
			name: "invalid HTTP",
			cfg: config.Config{
				HTTP: config.HTTPConfig{
					Port: -1,
				},
			},
			wantErr: "http.port must be between 0 and 65535",
		},
		{
			name: "invalid Tuya",
			cfg: config.Config{
				Tuya: config.TuyaConfig{
					ClientSecret: "client-secret",
				},
			},
			wantErr: "tuya.client_id is required",
		},
		{
			name: "invalid Yandex",
			cfg: config.Config{
				Tuya: config.TuyaConfig{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
				Yandex: config.YandexConfig{
					BearerToken: "bearer-token",
					PathPrefix:  config.DefaultYandexPathPrefix,
					OAuth: config.OAuthConfig{
						ClientID:     "oauth-client",
						ClientSecret: "oauth-secret",
					},
				},
			},
			wantErr: "yandex.user_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
