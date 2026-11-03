package model

// Client — данные клиента из подписанного client_id. Сервер их не хранит
type Client struct {
	Name         string   `json:"name"`          // Название клиента
	RedirectURIs []string `json:"redirect_uris"` // Разрешенные redirect_uri
}
