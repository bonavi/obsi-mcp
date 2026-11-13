package service

import (
	"net/url"
	"slices"
)

const (
	hostedCallback = "https://claude.ai/api/mcp/auth_callback" // Адрес возврата приложений Claude
)

// isAllowedRedirectURI пропускает адрес Claude и локальных клиентов: при открытой регистрации иначе код мог бы уйти чужому сайту
func (s *OAuthService) isAllowedRedirectURI(raw string) bool {
	return raw == hostedCallback || s.isLoopback(raw)
}

// isLoopback проверяет адрес возврата локального клиента: только localhost и только разрешённые пути
func (s *OAuthService) isLoopback(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "http" || !slices.Contains(s.config.LoopbackRedirectPaths, parsed.Path) || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return false
	}
	return parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1"
}

// sameRedirectURI сравнивает адреса. Порт loopback меняется от сессии к сессии, поэтому не учитывается
func (s *OAuthService) sameRedirectURI(registered, requested string) bool {
	if registered == requested {
		return true
	}
	if !s.isLoopback(registered) || !s.isLoopback(requested) {
		return false
	}
	first, _ := url.Parse(registered)
	second, _ := url.Parse(requested)
	return first.Hostname() == second.Hostname() && first.Path == second.Path
}
