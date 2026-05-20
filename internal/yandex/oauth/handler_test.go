package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOAuthAuthorizationCodeFlow(t *testing.T) {
	handler := newTestHandler()
	authorizeRequest := newAuthorizeRequest(
		"response_type=code",
		"client_id=smart-bridge",
		"redirect_uri="+testOAuthRedirectURI,
		"state=state-1",
		"scope=devices",
	)
	authorizeResponse := httptest.NewRecorder()

	handler.ServeHTTP(authorizeResponse, authorizeRequest)

	code := requireAuthorizationCode(t, authorizeResponse, "state-1")
	tokenRequest := newTokenRequest(
		"grant_type=authorization_code",
		"code="+code,
		"redirect_uri="+testOAuthRedirectURI,
	)
	tokenResponse := httptest.NewRecorder()

	handler.ServeHTTP(tokenResponse, tokenRequest)

	requireAccessTokenResponse(t, tokenResponse, "devices")
}

func TestOAuthAuthorizeValidatesRequest(t *testing.T) {
	tests := []struct {
		name   string
		params []string
	}{
		{
			name:   "unsupported response type",
			params: []string{"response_type=token", "client_id=smart-bridge", "redirect_uri=" + testOAuthRedirectURI},
		},
		{
			name:   "invalid client ID",
			params: []string{"response_type=code", "client_id=wrong", "redirect_uri=" + testOAuthRedirectURI},
		},
		{
			name:   "non HTTPS redirect URI",
			params: []string{"response_type=code", "client_id=smart-bridge", "redirect_uri=http://dialogs.yandex.ru/callback"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler()
			request := newAuthorizeRequest(tt.params...)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestOAuthTokenRefreshesAccessToken(t *testing.T) {
	handler := newTestHandler()
	refreshToken := newTestRefreshToken(t, handler.issuer)
	request := newTokenRequest("grant_type=refresh_token", "refresh_token="+refreshToken)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	requireAccessTokenResponse(t, response, "devices")
}

func TestOAuthTokenAcceptsFormClientCredentials(t *testing.T) {
	handler := newTestHandler()
	refreshToken := newTestRefreshToken(t, handler.issuer)
	request := newTokenRequestWithoutBasicAuth(
		"grant_type=refresh_token",
		"refresh_token="+refreshToken,
		"client_id=smart-bridge",
		"client_secret=client-secret",
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	requireAccessTokenResponse(t, response, "devices")
}

func TestOAuthTokenValidatesGrant(t *testing.T) {
	tests := []struct {
		name string
		body []string
		want string
	}{
		{
			name: "invalid authorization code",
			body: []string{"grant_type=authorization_code", "code=bad-code", "redirect_uri=" + testOAuthRedirectURI},
			want: `{"error":"invalid_grant","error_description":"invalid authorization code"}`,
		},
		{
			name: "invalid refresh token",
			body: []string{"grant_type=refresh_token", "refresh_token=bad-token"},
			want: `{"error":"invalid_grant","error_description":"invalid refresh token"}`,
		},
		{
			name: "unsupported grant type",
			body: []string{"grant_type=client_credentials"},
			want: `{"error":"unsupported_grant_type","error_description":"unsupported grant_type"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler()
			request := newTokenRequest(tt.body...)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.JSONEq(t, tt.want, response.Body.String())
		})
	}
}

func TestOAuthTokenRequiresBasicAuth(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		basicUser string
		basicPass string
	}{
		{
			name: "missing Basic credentials",
			body: "grant_type=refresh_token&refresh_token=token",
		},
		{
			name:      "wrong Basic secret",
			body:      "grant_type=refresh_token&refresh_token=token",
			basicUser: "smart-bridge",
			basicPass: "wrong-secret",
		},
		{
			name: "wrong form secret",
			body: "grant_type=refresh_token&refresh_token=token&client_id=smart-bridge&client_secret=wrong-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler()
			request := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.basicUser != "" || tt.basicPass != "" {
				request.SetBasicAuth(tt.basicUser, tt.basicPass)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			require.Equal(t, http.StatusUnauthorized, response.Code)
			require.Equal(t, `Basic realm="smart-bridge"`, response.Header().Get("WWW-Authenticate"))
			require.JSONEq(t, `{"error":"invalid_client","error_description":"invalid client credentials"}`, response.Body.String())
		})
	}
}
