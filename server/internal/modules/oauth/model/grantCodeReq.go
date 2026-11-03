package model

// GrantCodeReq — запрос кода после ввода пароля
type GrantCodeReq struct {
	Authorize AuthorizeReq // Параметры авторизации
	Password  string       // Пароль из формы
}
