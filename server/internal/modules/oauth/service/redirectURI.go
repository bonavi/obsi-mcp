package service

import (
	"net/url"
)

const (
	hostedCallback = "https://claude.ai/api/mcp/auth_callback" // Адрес возврата приложений Claude
	loopbackPath   = "/callback"                               // Путь возврата Claude Code
)

// isAllowedRedirectURI пропускает только адреса Claude: при открытой регистрации иначе код мог бы уйти чужому сайту
func isAllowedRedirectURI(raw string) bool {
	return raw == hostedCallback || isLoopback(raw)
}

// isLoopback проверяет адрес возврата Claude Code на локальной машине
func isLoopback(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" || parsed.Path != loopbackPath || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return false
	}
	return parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1"
}

// sameRedirectURI сравнивает адреса. Порт loopback меняется от сессии к сессии, поэтому не учитывается
func sameRedirectURI(registered, requested string) bool {
	if registered == requested {
		return true
	}
	if !isLoopback(registered) || !isLoopback(requested) {
		return false
	}
	first, _ := url.Parse(registered)
	second, _ := url.Parse(requested)
	return first.Hostname() == second.Hostname()
}
