package model

// OAuthError — ошибка протокола OAuth 2.1 (RFC 6749 §5.2)
type OAuthError struct {
	Code        string // Код для клиента
	Description string // Текст для разработчика клиента
}

// Error возвращает текст ошибки
func (e *OAuthError) Error() string {
	return e.Code + ": " + e.Description
}

// NewOAuthError создает ошибку протокола
func NewOAuthError(code, description string) *OAuthError {
	return &OAuthError{Code: code, Description: description}
}

// ErrInvalidCredentials — неверный пароль: клиенту не уходит, форма показывает
var ErrInvalidCredentials = NewOAuthError("access_denied", "invalid password")

// ErrTooManyAttempts — вход временно закрыт после неудачных попыток
var ErrTooManyAttempts = NewOAuthError("access_denied", "too many attempts")
