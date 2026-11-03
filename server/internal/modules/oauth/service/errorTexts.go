package service

// redirectNotRegistered — текст ошибки, которую нельзя отдавать редиректом
const redirectNotRegistered = "redirect_uri is not registered for this client"

// RedirectNotRegistered возвращает этот текст для эндпоинта
func RedirectNotRegistered() string {
	return redirectNotRegistered
}
