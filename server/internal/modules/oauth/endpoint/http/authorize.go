package http

import (
	stdErrors "errors"
	"net/http"
	"net/url"
	"server/internal/modules/oauth/model"
	oauthService "server/internal/modules/oauth/service"
)

// authorizeReq собирает параметры авторизации из query или полей формы
func authorizeReq(values url.Values) model.AuthorizeReq {
	return model.AuthorizeReq{
		ClientID:            values.Get("client_id"),
		RedirectURI:         values.Get("redirect_uri"),
		ResponseType:        values.Get("response_type"),
		CodeChallenge:       values.Get("code_challenge"),
		CodeChallengeMethod: values.Get("code_challenge_method"),
		Resource:            values.Get("resource"),
		State:               values.Get("state"),
	}
}

// AuthorizeForm проверяет запрос авторизации и показывает форму входа
func (e *OAuthEndpoint) AuthorizeForm(writer http.ResponseWriter, request *http.Request) {
	req := authorizeReq(request.URL.Query())
	client, err := e.oauthService.ValidateAuthorize(req)
	if err != nil {
		e.authorizeError(writer, request, req, err)
		return
	}
	renderForm(writer, http.StatusOK, newFormData(client, req, request.URL.Query(), ""))
}

// AuthorizeSubmit проверяет пароль и редиректит клиенту с кодом; при неверном пароле — форма снова
func (e *OAuthEndpoint) AuthorizeSubmit(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maximalFormBytes)
	if err := request.ParseForm(); err != nil {
		writeOAuthError(writer, model.NewOAuthError("invalid_request", "bad form"))
		return
	}
	req := authorizeReq(request.PostForm)

	code, err := e.oauthService.GrantCode(model.GrantCodeReq{Authorize: req, Password: request.PostForm.Get("password")})
	switch {
	case stdErrors.Is(err, model.ErrInvalidCredentials):
		e.rerenderForm(writer, request, req, http.StatusUnauthorized, "Неверный пароль")
		return
	case stdErrors.Is(err, model.ErrTooManyAttempts):
		e.rerenderForm(writer, request, req, http.StatusTooManyRequests, "Слишком много попыток, попробуйте позже")
		return
	case err != nil:
		e.authorizeError(writer, request, req, err)
		return
	}

	// Успех: код и state — клиенту
	redirect, err := url.Parse(req.RedirectURI)
	if err != nil {
		writeOAuthError(writer, err)
		return
	}
	query := redirect.Query()
	query.Set("code", code)
	if req.State != "" {
		query.Set("state", req.State)
	}
	redirect.RawQuery = query.Encode()
	http.Redirect(writer, request, redirect.String(), http.StatusFound)
}

// rerenderForm показывает форму с ошибкой входа
func (e *OAuthEndpoint) rerenderForm(writer http.ResponseWriter, request *http.Request, req model.AuthorizeReq, status int, text string) {
	client, err := e.oauthService.ValidateAuthorize(req)
	if err != nil {
		e.authorizeError(writer, request, req, err)
		return
	}
	renderForm(writer, status, newFormData(client, req, request.PostForm, text))
}

// authorizeError: ошибку клиента или адреса показываем пользователю, остальные — редиректом клиенту
func (e *OAuthEndpoint) authorizeError(writer http.ResponseWriter, request *http.Request, req model.AuthorizeReq, err error) {
	protocolError := oauthError(err)
	if protocolError.Code == "invalid_client" || protocolError.Code == "server_error" || protocolError.Description == oauthService.RedirectNotRegistered() {
		renderError(writer, http.StatusBadRequest, protocolError.Description)
		return
	}
	redirect, parseError := url.Parse(req.RedirectURI)
	if parseError != nil {
		renderError(writer, http.StatusBadRequest, protocolError.Description)
		return
	}
	query := redirect.Query()
	query.Set("error", protocolError.Code)
	query.Set("error_description", protocolError.Description)
	if req.State != "" {
		query.Set("state", req.State)
	}
	redirect.RawQuery = query.Encode()
	http.Redirect(writer, request, redirect.String(), http.StatusFound)
}
