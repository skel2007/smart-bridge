package oauth

type claims struct {
	Type        string `json:"typ"`
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri,omitempty"`
	Scope       string `json:"scope,omitempty"`
	ExpiresAt   int64  `json:"exp"`
	Nonce       string `json:"nonce"`
}
