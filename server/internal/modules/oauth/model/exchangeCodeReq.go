package model

// ExchangeCodeReq — обмен кода на токены
type ExchangeCodeReq struct {
	Code         string // Код авторизации
	ClientID     string // Подписанный client_id
	RedirectURI  string // Адрес возврата
	CodeVerifier string // PKCE verifier
	Resource     string // Канонический URL MCP
}
