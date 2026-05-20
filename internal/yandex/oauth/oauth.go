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
	ClientID          string
	ClientSecret      string
	StaticAccessToken string
}

const (
	codeTTL       = 5 * time.Minute
	tokenLifetime = 365 * 24 * time.Hour

	tokenTypeCode    = "code"
	tokenTypeRefresh = "refresh"
)

var errInvalidRedirectURI = errors.New("invalid redirect_uri")

type issuer struct {
	cfg   Config
	now   func() time.Time
	nonce func() (string, error)
}

type redirectRequest struct {
	RedirectURI string
	State       string
	Scope       string
}

func newIssuer(cfg Config) issuer {
	return issuer{
		cfg:   cfg,
		now:   time.Now,
		nonce: randomNonce,
	}
}

func (issuer issuer) newRedirectLocation(request redirectRequest) (string, error) {
	redirectURI, ok := validRedirectURI(request.RedirectURI)
	if !ok {
		return "", errInvalidRedirectURI
	}

	code, err := issuer.newToken(
		claims{
			Type:        tokenTypeCode,
			ClientID:    issuer.cfg.ClientID,
			RedirectURI: redirectURI.String(),
			Scope:       request.Scope,
		},
		codeTTL,
	)
	if err != nil {
		return "", err
	}

	out := redirectURI.Query()
	out.Set("code", code)
	if request.State != "" {
		out.Set("state", request.State)
	}
	redirectURI.RawQuery = out.Encode()

	return redirectURI.String(), nil
}

func (issuer issuer) exchangeAuthorizationCode(code string, redirectURI string) (string, bool) {
	tokenClaims, ok := issuer.parseToken(code, tokenTypeCode)
	if !ok || tokenClaims.RedirectURI == "" || redirectURI != tokenClaims.RedirectURI {
		return "", false
	}

	return tokenClaims.Scope, true
}

func (issuer issuer) exchangeRefreshToken(refreshToken string) (string, bool) {
	tokenClaims, ok := issuer.parseToken(refreshToken, tokenTypeRefresh)
	if !ok {
		return "", false
	}

	return tokenClaims.Scope, true
}

func (issuer issuer) newToken(tokenClaims claims, ttl time.Duration) (string, error) {
	nonce, err := issuer.nonce()
	if err != nil {
		return "", err
	}

	tokenClaims.ExpiresAt = issuer.now().Add(ttl).Unix()
	tokenClaims.Nonce = nonce

	payload, err := json.Marshal(tokenClaims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := issuer.signPayload(encodedPayload)

	return encodedPayload + "." + signature, nil
}

func (issuer issuer) parseToken(raw string, wantType string) (claims, bool) {
	encodedPayload, signature, ok := strings.Cut(raw, ".")
	if !ok {
		return claims{}, false
	}

	if !hmac.Equal([]byte(signature), []byte(issuer.signPayload(encodedPayload))) {
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

	if tokenClaims.Type != wantType || tokenClaims.ClientID != issuer.cfg.ClientID || issuer.now().Unix() > tokenClaims.ExpiresAt {
		return claims{}, false
	}

	return tokenClaims, true
}

func (issuer issuer) signPayload(encodedPayload string) string {
	mac := hmac.New(sha256.New, []byte(issuer.cfg.ClientSecret))
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
