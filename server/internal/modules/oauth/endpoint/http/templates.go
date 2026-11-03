package http

import (
	"html/template"
	"net/http"
	"net/url"
	"server/internal/modules/oauth/model"
)

// formData — данные формы входа
type formData struct {
	Client     string            // Название клиента, его присылает сам клиент
	Host       string            // Хост адреса возврата: куда уйдёт доступ
	IsLoopback bool              // Доступ получит программа на этом компьютере
	Hidden     map[string]string // Параметры авторизации для скрытых полей
	Error      string            // Текст ошибки входа
}

// hiddenParams — параметры авторизации, которые форма возвращает обратно
var hiddenParams = []string{"client_id", "redirect_uri", "response_type", "code_challenge", "code_challenge_method", "resource", "state"}

var formTemplate = template.Must(template.New("form").Parse(`<!doctype html>
<html lang="ru"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Obsidian — вход</title>
<style>
body{font:16px -apple-system,system-ui,sans-serif;background:#f2f2f7;margin:0;display:flex;min-height:100vh;align-items:center;justify-content:center}
form{background:#fff;padding:28px;border-radius:14px;width:320px;box-shadow:0 2px 12px rgba(0,0,0,.08)}
h1{font-size:20px;margin:0 0 6px}p{margin:0 0 18px;color:#555}
input{width:100%;box-sizing:border-box;padding:10px;margin:0 0 12px;border:1px solid #ccc;border-radius:8px;font-size:16px}
button{width:100%;padding:11px;border:0;border-radius:8px;background:#3e9bf0;color:#fff;font-size:16px}
.err{color:#c62828;margin:0 0 12px}
</style></head><body>
<form method="post" action="/oauth/authorize">
<h1>База Obsidian</h1>
<p><b>{{.Client}}</b> запрашивает чтение и изменение заметок. Доступ получит: <b>{{.Host}}</b>.</p>
{{if .IsLoopback}}<p class="err">Это программа на вашем компьютере. Разрешайте, только если сами сейчас её подключаете.</p>{{end}}
{{if .Error}}<p class="err">{{.Error}}</p>{{end}}
<input type="password" name="password" placeholder="Пароль" required autofocus autocomplete="current-password">
{{range $name, $value := .Hidden}}<input type="hidden" name="{{$name}}" value="{{$value}}">{{end}}
<button type="submit">Разрешить</button>
</form></body></html>`))

var errorTemplate = template.Must(template.New("error").Parse(`<!doctype html>
<html lang="ru"><head><meta charset="utf-8"><title>Obsidian — ошибка</title>
<style>body{font:16px -apple-system,system-ui,sans-serif;padding:40px;color:#333}</style></head>
<body><h1>Не удалось авторизовать клиента</h1><p>{{.}}</p></body></html>`))

// newFormData собирает данные формы
func newFormData(client model.Client, req model.AuthorizeReq, params url.Values, text string) formData {
	hidden := make(map[string]string, len(hiddenParams))
	for _, name := range hiddenParams {
		if value := params.Get(name); value != "" {
			hidden[name] = value
		}
	}
	host := ""
	if parsed, err := url.Parse(req.RedirectURI); err == nil {
		host = parsed.Hostname()
	}
	return formData{Client: client.Name, Host: host, IsLoopback: host == "localhost" || host == "127.0.0.1", Hidden: hidden, Error: text}
}

// setPageHeaders запрещает кэш, встраивание страницы в чужой сайт и передачу адреса дальше
func setPageHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Frame-Options", "DENY")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'")
}

// renderForm показывает форму входа
func renderForm(writer http.ResponseWriter, status int, data formData) {
	setPageHeaders(writer)
	writer.WriteHeader(status)
	_ = formTemplate.Execute(writer, data)
}

// renderError показывает ошибку, которую нельзя отдать клиенту редиректом
func renderError(writer http.ResponseWriter, status int, text string) {
	setPageHeaders(writer)
	writer.WriteHeader(status)
	_ = errorTemplate.Execute(writer, text)
}
