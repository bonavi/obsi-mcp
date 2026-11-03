package model

// AuthorizeReq — параметры запроса авторизации (RFC 6749 §4.1.1 + PKCE + resource)
type AuthorizeReq struct {
	ClientID            string // Подписанный client_id
	RedirectURI         string // Адрес возврата, должен быть у клиента
	ResponseType        string // Только code
	CodeChallenge       string // PKCE challenge
	CodeChallengeMethod string // Только S256
	Resource            string // Канонический URL MCP, пусто — подставляется свой
	State               string // Возвращается клиенту как есть
}
