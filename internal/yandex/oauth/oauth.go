package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	ClientID     string
	ClientSecret string
}

const (
	codeTTL       = 5 * time.Minute
	tokenLifetime = 365 * 24 * time.Hour

	tokenTypeCode    = "code"
	tokenTypeRefresh = "refresh"
)

var errInvalidRedirectURI = errors.New("invalid redirect_uri")

func newRedirectLocation(cfg Config, now func() time.Time, rawRedirectURI string, state string, scope string) (string, error) {
	redirectURI, ok := validRedirectURI(rawRedirectURI)
	if !ok {
		return "", errInvalidRedirectURI
	}

	code, err := newToken(
		cfg,
		now,
		claims{
			Type:        tokenTypeCode,
			ClientID:    cfg.ClientID,
			RedirectURI: redirectURI.String(),
			Scope:       scope,
		},
		codeTTL,
	)
	if err != nil {
		return "", err
	}

	out := redirectURI.Query()
	out.Set("code", code)
	if state != "" {
		out.Set("state", state)
	}
	redirectURI.RawQuery = out.Encode()

	return redirectURI.String(), nil
}

func exchangeAuthorizationCode(cfg Config, now func() time.Time, code string, redirectURI string) (string, bool) {
	tokenClaims, ok := parseToken(cfg, now, code, tokenTypeCode)
	if !ok || tokenClaims.RedirectURI == "" || redirectURI != tokenClaims.RedirectURI {
		return "", false
	}

	return tokenClaims.Scope, true
}

func exchangeRefreshToken(cfg Config, now func() time.Time, refreshToken string) (string, bool) {
	tokenClaims, ok := parseToken(cfg, now, refreshToken, tokenTypeRefresh)
	if !ok {
		return "", false
	}

	return tokenClaims.Scope, true
}

func newToken(cfg Config, now func() time.Time, tokenClaims claims, ttl time.Duration) (string, error) {
	nonce, err := randomNonce()
	if err != nil {
		return "", err
	}

	tokenClaims.ExpiresAt = now().Add(ttl).Unix()
	tokenClaims.Nonce = nonce

	payload, err := json.Marshal(tokenClaims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := signPayload(cfg, encodedPayload)

	return encodedPayload + "." + signature, nil
}

func parseToken(cfg Config, now func() time.Time, raw string, wantType string) (claims, bool) {
	encodedPayload, signature, ok := strings.Cut(raw, ".")
	if !ok {
		return claims{}, false
	}

	if !hmac.Equal([]byte(signature), []byte(signPayload(cfg, encodedPayload))) {
		return claims{}, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return claims{}, false
	}

	var tokenClaims claims
	if err := json.Unmarshal(payload, &tokenClaims); err != nil {
		return claims{}, false
	}

	if tokenClaims.Type != wantType || tokenClaims.ClientID != cfg.ClientID || now().Unix() > tokenClaims.ExpiresAt {
		return claims{}, false
	}

	return tokenClaims, true
}

func signPayload(cfg Config, encodedPayload string) string {
	mac := hmac.New(sha256.New, []byte(cfg.ClientSecret))
	_, _ = mac.Write([]byte(encodedPayload))

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validRedirectURI(raw string) (*url.URL, bool) {
	redirectURI, err := url.Parse(raw)
	if err != nil || redirectURI.Scheme != "https" || redirectURI.Host == "" {
		return nil, false
	}

	return redirectURI, true
}

func randomNonce() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}
