package oauth

import (
	"crypto/hmac"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Handler serves OAuth compatibility endpoints.
type Handler struct {
	cfg    Config
	issuer issuer
	logger *slog.Logger
	mux    *http.ServeMux
}

type Option func(*Handler)

func WithLogger(logger *slog.Logger) Option {
	return func(handler *Handler) {
		if logger != nil {
			handler.logger = logger
		}
	}
}

func NewHandler(cfg Config, options ...Option) *Handler {
	handler := &Handler{
		cfg:    cfg,
		issuer: newIssuer(cfg),
		logger: slog.New(slog.DiscardHandler),
		mux:    http.NewServeMux(),
	}
	for _, option := range options {
		option(handler)
	}

	handler.mux.HandleFunc("GET /authorize", handler.serveOAuthAuthorize)
	handler.mux.HandleFunc("POST /token", handler.serveOAuthToken)

	return handler
}

func (handler *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.mux.ServeHTTP(w, r)
}

func (handler *Handler) serveOAuthAuthorize(w http.ResponseWriter, r *http.Request) {
	request, ok := newOAuthAuthorizeRequest(r)
	if !ok {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	if !handler.oauthClientIDMatches(request.ClientID) {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return
	}

	redirectLocation, err := handler.issuer.newRedirectLocation(request.Redirect)
	if err != nil {
		if errors.Is(err, errInvalidRedirectURI) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "create authorization code", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, redirectLocation, http.StatusFound)
}

type oauthAuthorizeRequest struct {
	ClientID string
	Redirect redirectRequest
}

func newOAuthAuthorizeRequest(r *http.Request) (oauthAuthorizeRequest, bool) {
	query := r.URL.Query()
	if query.Get("response_type") != "code" {
		return oauthAuthorizeRequest{}, false
	}

	return oauthAuthorizeRequest{
		ClientID: query.Get("client_id"),
		Redirect: redirectRequest{
			RedirectURI: query.Get("redirect_uri"),
			State:       query.Get("state"),
			Scope:       query.Get("scope"),
		},
	}, true
}

func (handler *Handler) serveOAuthToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "invalid form body")
		return
	}
	if !handler.oauthClientAuthorized(r) {
		handler.logTokenError(r, "invalid_client", "oauth token client rejected")
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "invalid client credentials")
		return
	}

	switch r.Form.Get("grant_type") {
	case "authorization_code":
		handler.serveOAuthAuthorizationCode(w, r)
	case "refresh_token":
		handler.serveOAuthRefreshToken(w, r)
	default:
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "unsupported grant_type")
	}
}

func (handler *Handler) serveOAuthAuthorizationCode(w http.ResponseWriter, r *http.Request) {
	scope, ok := handler.issuer.exchangeAuthorizationCode(
		r.Form.Get("code"),
		r.Form.Get("redirect_uri"),
	)
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "invalid authorization code")
		return
	}

	handler.writeOAuthTokenResponse(w, scope)
}

func (handler *Handler) serveOAuthRefreshToken(w http.ResponseWriter, r *http.Request) {
	scope, ok := handler.issuer.exchangeRefreshToken(
		r.Form.Get("refresh_token"),
	)
	if !ok {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "invalid refresh token")
		return
	}

	handler.writeOAuthTokenResponse(w, scope)
}

func (handler *Handler) writeOAuthTokenResponse(w http.ResponseWriter, scope string) {
	refreshToken, err := handler.issuer.newToken(
		claims{
			Type:     tokenTypeRefresh,
			ClientID: handler.cfg.ClientID,
			Scope:    scope,
		},
		tokenLifetime,
	)
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "create refresh token")
		return
	}

	writeOAuthJSON(w, http.StatusOK, oauthTokenResponse{
		AccessToken:  handler.cfg.StaticAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(tokenLifetime.Seconds()),
		RefreshToken: refreshToken,
		Scope:        scope,
	})
}

func (handler *Handler) oauthClientAuthorized(r *http.Request) bool {
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientID = r.Form.Get("client_id")
		clientSecret = r.Form.Get("client_secret")
	}

	return handler.oauthClientIDMatches(clientID) && hmac.Equal([]byte(clientSecret), []byte(handler.cfg.ClientSecret))
}

func (handler *Handler) logTokenError(r *http.Request, code string, message string) {
	handler.logger.WarnContext(r.Context(), message,
		"error_code", code,
		"grant_type", r.Form.Get("grant_type"),
	)
}

func (handler *Handler) oauthClientIDMatches(clientID string) bool {
	return handler.cfg.ClientID != "" && hmac.Equal([]byte(clientID), []byte(handler.cfg.ClientID))
}

func writeOAuthError(w http.ResponseWriter, status int, code string, description string) {
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Basic realm="smart-bridge"`)
	}

	writeOAuthJSON(w, status, oauthErrorResponse{
		Error:            code,
		ErrorDescription: description,
	})
}

func writeOAuthJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	writeJSON(w, status, value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
