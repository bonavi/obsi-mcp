package mcp

import "net/http"

// statusRecorder запоминает код ответа
type statusRecorder struct {
	http.ResponseWriter     // Исходный ответ
	status              int // Код ответа
}

// WriteHeader запоминает код и передаёт его дальше
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
