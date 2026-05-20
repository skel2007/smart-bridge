package oauth

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testOAuthRedirectURI = "https://dialogs.yandex.ru/callback"

var testOAuthNow = time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

func TestNewRedirectLocationSignsCodeAndPreservesState(t *testing.T) {
	cfg := testConfig()

	location, err := newRedirectLocation(cfg, testNow, testOAuthRedirectURI, "state-1", "devices")

	require.NoError(t, err)
	redirectURI, err := url.Parse(location)
	require.NoError(t, err)
	query := redirectURI.Query()

	require.Equal(t, "state-1", query.Get("state"))
	code := query.Get("code")
	require.NotEmpty(t, code)

	expectedLocation, err := url.Parse(testOAuthRedirectURI)
	require.NoError(t, err)
	expectedLocation.RawQuery = url.Values{
		"code":  []string{code},
		"state": []string{"state-1"},
	}.Encode()
	require.Equal(t, expectedLocation.String(), location)

	tokenClaims, ok := parseToken(cfg, testNow, code, tokenTypeCode)
	require.True(t, ok)
	require.Equal(t, cfg.ClientID, tokenClaims.ClientID)
	require.Equal(t, testOAuthRedirectURI, tokenClaims.RedirectURI)
	require.Equal(t, "devices", tokenClaims.Scope)
	require.Equal(t, testOAuthNow.Add(codeTTL).Unix(), tokenClaims.ExpiresAt)
}

func TestNewRedirectLocationRejectsInvalidRedirectURI(t *testing.T) {
	cfg := testConfig()

	_, err := newRedirectLocation(cfg, testNow, "http://dialogs.yandex.ru/callback", "", "")

	require.ErrorIs(t, err, errInvalidRedirectURI)
}

func TestExchangeAuthorizationCodeRejectsRedirectURIMismatch(t *testing.T) {
	cfg := testConfig()
	code := newTestOAuthCode(t, cfg, testNow, testOAuthRedirectURI, "")

	_, ok := exchangeAuthorizationCode(cfg, testNow, code, "https://example.com/callback")

	require.False(t, ok)
}

func TestExchangeAuthorizationCodeRejectsExpiredCode(t *testing.T) {
	current := testOAuthNow
	now := func() time.Time { return current }
	cfg := testConfig()
	code := newTestOAuthCode(t, cfg, now, testOAuthRedirectURI, "")
	current = current.Add(codeTTL + time.Second)

	_, ok := exchangeAuthorizationCode(cfg, now, code, testOAuthRedirectURI)

	require.False(t, ok)
}

func TestExchangeRefreshTokenReturnsScope(t *testing.T) {
	cfg := testConfig()
	refreshToken, err := newToken(
		cfg,
		testNow,
		claims{
			Type:     tokenTypeRefresh,
			ClientID: cfg.ClientID,
			Scope:    "devices",
		},
		tokenLifetime,
	)
	require.NoError(t, err)

	scope, ok := exchangeRefreshToken(cfg, testNow, refreshToken)

	require.True(t, ok)
	require.Equal(t, "devices", scope)
}

func testNow() time.Time {
	return testOAuthNow
}

func testConfig() Config {
	return Config{
		ClientID:     "smart-bridge",
		ClientSecret: "client-secret",
	}
}

func newTestOAuthCode(t *testing.T, cfg Config, now func() time.Time, redirectURI string, scope string) string {
	t.Helper()

	code, err := newToken(
		cfg,
		now,
		claims{
			Type:        tokenTypeCode,
			ClientID:    cfg.ClientID,
			RedirectURI: redirectURI,
			Scope:       scope,
		},
		codeTTL,
	)
	require.NoError(t, err)

	return code
}
