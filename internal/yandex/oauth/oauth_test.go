package oauth

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewRedirectLocationReturnsCodeAndState(t *testing.T) {
	issuer := testIssuer()

	location, err := issuer.newRedirectLocation(redirectRequest{
		RedirectURI: testOAuthRedirectURI,
		State:       "state-1",
	})

	require.NoError(t, err)
	redirectURL, err := url.Parse(location)
	require.NoError(t, err)
	query := redirectURL.Query()

	require.Equal(t, testOAuthRedirectURI, baseURI(redirectURL))
	require.Equal(t, "state-1", query.Get("state"))
	require.NotEmpty(t, query.Get("code"))
}

func TestNewRedirectLocationRejectsNonHTTPSRedirectURI(t *testing.T) {
	issuer := testIssuer()

	_, err := issuer.newRedirectLocation(redirectRequest{
		RedirectURI: "http://dialogs.yandex.ru/callback",
	})

	require.ErrorIs(t, err, errInvalidRedirectURI)
}

func TestExchangeAuthorizationCode(t *testing.T) {
	tests := []struct {
		name          string
		redirectURI   string
		advanceTimeBy time.Duration
		wantScope     string
		wantOK        bool
	}{
		{
			name:        "returns scope",
			redirectURI: testOAuthRedirectURI,
			wantScope:   "devices",
			wantOK:      true,
		},
		{
			name:        "rejects redirect URI mismatch",
			redirectURI: "https://example.com/callback",
		},
		{
			name:          "rejects expired code",
			redirectURI:   testOAuthRedirectURI,
			advanceTimeBy: codeTTL + time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuer := testIssuer()
			code := newTestAuthorizationCode(t, issuer)
			if tt.advanceTimeBy != 0 {
				issuer.now = testTimeAfter(tt.advanceTimeBy)
			}

			scope, ok := issuer.exchangeAuthorizationCode(code, tt.redirectURI)

			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantScope, scope)
		})
	}
}

func TestExchangeAuthorizationCodeRejectsRefreshToken(t *testing.T) {
	issuer := testIssuer()
	refreshToken := newTestRefreshToken(t, issuer)

	_, ok := issuer.exchangeAuthorizationCode(refreshToken, testOAuthRedirectURI)

	require.False(t, ok)
}

func TestExchangeRefreshToken(t *testing.T) {
	tests := []struct {
		name          string
		advanceTimeBy time.Duration
		wantScope     string
		wantOK        bool
	}{
		{
			name:      "returns scope",
			wantScope: "devices",
			wantOK:    true,
		},
		{
			name:          "rejects expired refresh token",
			advanceTimeBy: tokenLifetime + time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issuer := testIssuer()
			refreshToken := newTestRefreshToken(t, issuer)
			if tt.advanceTimeBy != 0 {
				issuer.now = testTimeAfter(tt.advanceTimeBy)
			}

			scope, ok := issuer.exchangeRefreshToken(refreshToken)

			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantScope, scope)
		})
	}
}

func TestExchangeRefreshTokenRejectsAuthorizationCode(t *testing.T) {
	issuer := testIssuer()
	code := newTestAuthorizationCode(t, issuer)

	_, ok := issuer.exchangeRefreshToken(code)

	require.False(t, ok)
}
