package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testOAuthRedirectURI = "https://dialogs.yandex.ru/callback"

var testOAuthNow = time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

func testIssuer() issuer {
	issuer := newIssuer(testConfig())
	issuer.now = testTimeAt(testOAuthNow)
	issuer.nonce = testNonce

	return issuer
}

func newTestHandler() *Handler {
	handler := NewHandler(testConfig())
	handler.issuer.now = testTimeAt(testOAuthNow)
	handler.issuer.nonce = testNonce

	return handler
}

func testTimeAt(now time.Time) func() time.Time {
	return func() time.Time { return now }
}

func testTimeAfter(delta time.Duration) func() time.Time {
	return testTimeAt(testOAuthNow.Add(delta))
}

func testNonce() (string, error) {
	return "nonce", nil
}

func testConfig() Config {
	return Config{
		ClientID:          "smart-bridge",
		ClientSecret:      "client-secret",
		StaticAccessToken: "secret-token",
	}
}

func newTestAuthorizationCode(t *testing.T, issuer issuer) string {
	t.Helper()

	code, err := issuer.newToken(
		claims{
			Type:        tokenTypeCode,
			ClientID:    issuer.cfg.ClientID,
			RedirectURI: testOAuthRedirectURI,
			Scope:       "devices",
		},
		codeTTL,
	)
	require.NoError(t, err)

	return code
}

func newTestRefreshToken(t *testing.T, issuer issuer) string {
	t.Helper()

	refreshToken, err := issuer.newToken(
		claims{
			Type:     tokenTypeRefresh,
			ClientID: issuer.cfg.ClientID,
			Scope:    "devices",
		},
		tokenLifetime,
	)
	require.NoError(t, err)

	return refreshToken
}

func newAuthorizeRequest(params ...string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/authorize?"+encodeForm(params...), nil)
}

func newTokenRequest(params ...string) *http.Request {
	request := newTokenRequestWithoutBasicAuth(params...)
	request.SetBasicAuth("smart-bridge", "client-secret")

	return request
}

func newTokenRequestWithoutBasicAuth(params ...string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(encodeForm(params...)))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return request
}

func encodeForm(params ...string) string {
	values := make(url.Values, len(params))
	for _, param := range params {
		name, value, _ := strings.Cut(param, "=")
		values.Set(name, value)
	}

	return values.Encode()
}

func requireAuthorizationCode(t *testing.T, response *httptest.ResponseRecorder, state string) string {
	t.Helper()

	require.Equal(t, http.StatusFound, response.Code)
	location, err := url.Parse(response.Header().Get("Location"))
	require.NoError(t, err)
	require.Equal(t, testOAuthRedirectURI, baseURI(location))
	require.Equal(t, state, location.Query().Get("state"))

	code := location.Query().Get("code")
	require.NotEmpty(t, code)

	return code
}

func requireAccessTokenResponse(t *testing.T, response *httptest.ResponseRecorder, scope string) {
	t.Helper()

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", response.Header().Get("Pragma"))

	var out oauthTokenResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &out))
	require.Equal(t, "secret-token", out.AccessToken)
	require.Equal(t, "Bearer", out.TokenType)
	require.Equal(t, int64(tokenLifetime.Seconds()), out.ExpiresIn)
	require.Equal(t, scope, out.Scope)
	require.NotEmpty(t, out.RefreshToken)
}

func baseURI(u *url.URL) string {
	return u.Scheme + "://" + u.Host + u.Path
}
