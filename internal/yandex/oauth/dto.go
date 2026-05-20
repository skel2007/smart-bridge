package oauth

type claims struct {
	Type        string `json:"typ"`
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	Scope       string `json:"scope,omitempty"`
	ExpiresAt   int64  `json:"exp"`
	Nonce       string `json:"nonce"`
}

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
