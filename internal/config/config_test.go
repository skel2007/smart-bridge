package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	path := writeConfig(t, `
tuya:
  endpoint: https://example.com
  client_id: client-id
  client_secret: client-secret
`)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "https://example.com", cfg.Tuya.Endpoint)
	require.Equal(t, "client-id", cfg.Tuya.ClientID)
	require.Equal(t, "client-secret", cfg.Tuya.ClientSecret)
}

func TestLoadAppliesDefaultTuyaEndpoint(t *testing.T) {
	path := writeConfig(t, `
tuya:
  client_id: client-id
  client_secret: client-secret
`)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, DefaultTuyaEndpoint, cfg.Tuya.Endpoint)
}

func TestLoadReturnsClearErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "invalid YAML",
			content: `tuya: [`,
			wantErr: "invalid YAML in config file",
		},
		{
			name: "missing client ID",
			content: `
tuya:
  client_secret: client-secret
`,
			wantErr: "tuya.client_id is required",
		},
		{
			name: "missing client secret",
			content: `
tuya:
  client_id: client-id
`,
			wantErr: "tuya.client_secret is required",
		},
		{
			name: "does not leak client secret",
			content: `
tuya:
  client_id: ""
  client_secret: super-secret
`,
			wantErr: "tuya.client_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.content)

			_, err := Load(path)
			require.Error(t, err)
			require.ErrorContains(t, err, tt.wantErr)
			require.NotContains(t, err.Error(), "super-secret")
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := Load(path)
	require.EqualError(t, err, "config file not found: "+path)
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
