package integrationTest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/crypto/argon2"

	"pkg/jwtManager"
	oauthEndpointHTTP "server/internal/modules/oauth/endpoint/http"
	oauthModel "server/internal/modules/oauth/model"
	oauthRepository "server/internal/modules/oauth/repository"
	oauthService "server/internal/modules/oauth/service"
	mcpTransport "server/internal/utils/mcp"
)

const (
	testPassword = "correct horse battery staple"            // Пароль в тестах
	testVerifier = "test-code-verifier-0123456789"           // PKCE verifier в тестах
	testRedirect = "https://claude.ai/api/mcp/auth_callback" // Адрес возврата в тестах
)

// authServer — сервер авторизации в тесте
type authServer struct {
	url       string // Адрес сервера
	stateFile string // Файл сессий
}

// noRedirectClient не следует за редиректами: нужен их адрес
var noRedirectClient = &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }} //nolint:exhaustruct

// passwordHash считает хэш пароля с минимально допустимыми параметрами
func passwordHash() string {
	salt := []byte("0123456789abcdef")
	hash := argon2.IDKey([]byte(testPassword), salt, 1, 32*1024, 1, 32)
	return "$argon2id$v=19$m=32768,t=1,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(hash)
}

// newAuthServer поднимает MCP вместе с OAuth, как в main.go
func newAuthServer(t *testing.T, server *sdk.Server) (*httptest.Server, *authServer) {
	t.Helper()
	jwtManager.Init([]byte("test-signing-key-test-signing-key"), 30*time.Minute, 720*time.Hour)
	stateFile := filepath.Join(t.TempDir(), "state.json")
	mux := http.NewServeMux()
	testServer := httptest.NewUnstartedServer(mux)
	testServer.Start()
	mountOAuth(t, mux, server, stateFile, testServer.URL)
	return testServer, &authServer{url: testServer.URL, stateFile: stateFile}
}

// mountOAuth подключает маршруты OAuth и MCP
func mountOAuth(t *testing.T, mux *http.ServeMux, server *sdk.Server, stateFile, publicURL string) {
	t.Helper()
	repository, err := oauthRepository.NewOAuthRepository(stateFile)
	if err != nil {
		t.Fatal(err)
	}
	service, err := oauthService.NewOAuthService(repository, oauthService.Config{
		PublicURL: publicURL, AccessTokenTTL: 30 * time.Minute, RefreshTokenTTL: 720 * time.Hour,
		PasswordHash: passwordHash(), ClientKey: []byte("client-key"),
	})
	if err != nil {
		t.Fatal(err)
	}
	oauthEndpointHTTP.NewOAuthEndpoint(service).Register(mux)
	mux.Handle("GET "+mcpTransport.ResourceMetadataPath, mcpTransport.NewResourceMetadataHandler(publicURL))
	mux.Handle("/", mcpTransport.NewHandler(server, service, publicURL))
}

// register регистрирует клиента и возвращает client_id
func (a *authServer) register(t *testing.T, redirect string) (int, string) {
	t.Helper()
	body, _ := json.Marshal(oauthModel.RegisterClientReq{ClientName: "Тест", RedirectURIs: []string{redirect}})
	response, err := http.Post(a.url+"/oauth/register", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var res oauthModel.RegisterClientRes
	json.NewDecoder(response.Body).Decode(&res)
	return response.StatusCode, res.ClientID
}

// authorize отправляет форму входа
func (a *authServer) authorize(t *testing.T, clientID, redirect, password string, change func(url.Values)) *http.Response {
	t.Helper()
	sum := sha256.Sum256([]byte(testVerifier))
	form := url.Values{
		"client_id": {clientID}, "redirect_uri": {redirect}, "response_type": {"code"},
		"code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"},
		"resource": {a.url}, "state": {"xyz"}, "password": {password},
	}
	if change != nil {
		change(form)
	}
	response, err := noRedirectClient.PostForm(a.url+"/oauth/authorize", form)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	return response
}

// code проходит вход и возвращает client_id и код
func (a *authServer) code(t *testing.T) (string, string) {
	t.Helper()
	_, clientID := a.register(t, testRedirect)
	response := a.authorize(t, clientID, testRedirect, testPassword, nil)
	if response.StatusCode != http.StatusFound {
		t.Fatalf("authorize status %d", response.StatusCode)
	}
	location, err := url.Parse(response.Header.Get("Location"))
	if err != nil || location.Query().Get("state") != "xyz" {
		t.Fatalf("location %v", response.Header.Get("Location"))
	}
	return clientID, location.Query().Get("code")
}

// token вызывает эндпоинт токенов
func (a *authServer) token(t *testing.T, form url.Values) (int, oauthModel.TokenRes, string) {
	t.Helper()
	response, err := http.PostForm(a.url+"/oauth/token", form)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	var res oauthModel.TokenRes
	json.Unmarshal(body, &res)
	return response.StatusCode, res, string(body)
}

// exchangeForm собирает обмен кода
func exchangeForm(clientID, code string) url.Values {
	return url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {clientID}, "redirect_uri": {testRedirect}, "code_verifier": {testVerifier}}
}

// login проходит весь вход и возвращает токены
func (a *authServer) login(t *testing.T) oauthModel.TokenRes {
	t.Helper()
	clientID, code := a.code(t)
	status, res, body := a.token(t, exchangeForm(clientID, code))
	if status != http.StatusOK || res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatalf("token: %d %s", status, body)
	}
	return res
}

// bareAuth поднимает OAuth с пустым MCP-сервером
func bareAuth(t *testing.T) *authServer {
	t.Helper()
	testServer, auth := newAuthServer(t, mcpTransport.NewServer("test"))
	t.Cleanup(testServer.Close)
	return auth
}

// mcpStatus возвращает код ответа MCP на токен
func (a *authServer) mcpStatus(t *testing.T, token string) int {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, a.url+"/", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	return response.StatusCode
}

func TestDiscovery(t *testing.T) {
	auth := bareAuth(t)
	response, err := http.Post(auth.url+"/", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized || !strings.Contains(response.Header.Get("WWW-Authenticate"), mcpTransport.ResourceMetadataPath) {
		t.Fatalf("401 without pointer: %d %q", response.StatusCode, response.Header.Get("WWW-Authenticate"))
	}
	metadata, err := http.Get(auth.url + "/.well-known/oauth-authorization-server")
	if err != nil {
		t.Fatal(err)
	}
	defer metadata.Body.Close()
	body, _ := io.ReadAll(metadata.Body)
	if !strings.Contains(string(body), `"S256"`) || !strings.Contains(string(body), "/oauth/register") {
		t.Fatalf("metadata %s", body)
	}
}

func TestRegisterRejectsForeignRedirects(t *testing.T) {
	auth := bareAuth(t)
	for _, redirect := range []string{"https://evil.example/callback", "http://claude.ai/api/mcp/auth_callback", "https://claude.ai.evil.example/api/mcp/auth_callback", "http://localhost:1234/other", "http://user@localhost:1234/callback", "javascript:alert(1)"} {
		if status, _ := auth.register(t, redirect); status != http.StatusBadRequest {
			t.Errorf("%s: status %d", redirect, status)
		}
	}
	for _, redirect := range []string{testRedirect, "http://localhost:3118/callback", "http://127.0.0.1:9/callback"} {
		if status, _ := auth.register(t, redirect); status != http.StatusCreated {
			t.Errorf("%s: status %d", redirect, status)
		}
	}
}

func TestAuthorizeRejectsTampering(t *testing.T) {
	auth := bareAuth(t)
	_, clientID := auth.register(t, testRedirect)

	// Подпись client_id защищает данные клиента
	if response := auth.authorize(t, clientID+"x", testRedirect, testPassword, nil); response.StatusCode != http.StatusBadRequest {
		t.Errorf("forged client: %d", response.StatusCode)
	}

	// Чужой адрес возврата не получает редирект
	if response := auth.authorize(t, clientID, "https://evil.example/cb", testPassword, nil); response.StatusCode != http.StatusBadRequest || response.Header.Get("Location") != "" {
		t.Errorf("foreign redirect: %d %q", response.StatusCode, response.Header.Get("Location"))
	}

	// Без PKCE кода нет
	response := auth.authorize(t, clientID, testRedirect, testPassword, func(form url.Values) { form.Del("code_challenge") })
	if strings.Contains(response.Header.Get("Location"), "code=") {
		t.Error("code issued without PKCE")
	}

	// Чужой resource
	response = auth.authorize(t, clientID, testRedirect, testPassword, func(form url.Values) { form.Set("resource", "https://other.example") })
	if !strings.Contains(response.Header.Get("Location"), "invalid_target") {
		t.Errorf("foreign resource: %q", response.Header.Get("Location"))
	}
}

func TestLoopbackPortIsIgnored(t *testing.T) {
	auth := bareAuth(t)
	_, clientID := auth.register(t, "http://localhost/callback")
	if response := auth.authorize(t, clientID, "http://localhost:3118/callback", testPassword, nil); response.StatusCode != http.StatusFound {
		t.Errorf("status %d", response.StatusCode)
	}
	if response := auth.authorize(t, clientID, "http://127.0.0.1:3118/callback", testPassword, nil); response.StatusCode != http.StatusBadRequest {
		t.Errorf("other host: status %d", response.StatusCode)
	}
}

func TestPasswordLockout(t *testing.T) {
	auth := bareAuth(t)
	_, clientID := auth.register(t, testRedirect)
	for i := range 5 {
		if response := auth.authorize(t, clientID, testRedirect, "wrong", nil); response.StatusCode != http.StatusUnauthorized {
			t.Fatalf("attempt %d: %d", i, response.StatusCode)
		}
	}

	// После блокировки не проходит даже верный пароль
	if response := auth.authorize(t, clientID, testRedirect, testPassword, nil); response.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("locked: %d", response.StatusCode)
	}
}

func TestCodeRules(t *testing.T) {
	auth := bareAuth(t)

	// Код одноразовый
	clientID, code := auth.code(t)
	if status, _, body := auth.token(t, exchangeForm(clientID, code)); status != http.StatusOK {
		t.Fatalf("first exchange: %s", body)
	}
	if status, _, _ := auth.token(t, exchangeForm(clientID, code)); status != http.StatusBadRequest {
		t.Errorf("second exchange: %d", status)
	}

	// Неверный verifier сжигает код
	clientID, code = auth.code(t)
	form := exchangeForm(clientID, code)
	form.Set("code_verifier", "wrong")
	if status, _, _ := auth.token(t, form); status != http.StatusBadRequest {
		t.Errorf("bad verifier: %d", status)
	}
	if status, _, _ := auth.token(t, exchangeForm(clientID, code)); status != http.StatusBadRequest {
		t.Errorf("code survived a bad verifier: %d", status)
	}

	// Другой redirect_uri и resource
	clientID, code = auth.code(t)
	form = exchangeForm(clientID, code)
	form.Set("redirect_uri", "http://localhost:1/callback")
	if status, _, _ := auth.token(t, form); status != http.StatusBadRequest {
		t.Errorf("bad redirect: %d", status)
	}
	clientID, code = auth.code(t)
	form = exchangeForm(clientID, code)
	form.Set("resource", "https://other.example")
	if status, _, _ := auth.token(t, form); status != http.StatusBadRequest {
		t.Errorf("bad resource: %d", status)
	}
}

func TestRefreshRotationAndTheft(t *testing.T) {
	auth := bareAuth(t)
	first := auth.login(t)
	if status := auth.mcpStatus(t, first.AccessToken); status != http.StatusOK {
		t.Fatalf("valid token: %d", status)
	}

	// Ротация выдаёт новый токен
	refresh := func(token string) (int, oauthModel.TokenRes) {
		status, res, _ := auth.token(t, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {token}})
		return status, res
	}
	status, second := refresh(first.RefreshToken)
	if status != http.StatusOK || second.RefreshToken == first.RefreshToken {
		t.Fatalf("rotation: %d", status)
	}

	// Повтор старого токена отзывает всю цепочку, включая действующий access
	if status, _ = refresh(first.RefreshToken); status != http.StatusBadRequest {
		t.Fatalf("reuse: %d", status)
	}
	if status, _ = refresh(second.RefreshToken); status != http.StatusBadRequest {
		t.Errorf("family survived: %d", status)
	}
	if status = auth.mcpStatus(t, second.AccessToken); status != http.StatusUnauthorized {
		t.Errorf("access survived revocation: %d", status)
	}
}

func TestRefreshTokenIsNotAccessToken(t *testing.T) {
	auth := bareAuth(t)
	tokens := auth.login(t)
	for _, token := range []string{tokens.RefreshToken, "garbage", ""} {
		if status := auth.mcpStatus(t, token); status != http.StatusUnauthorized {
			t.Errorf("token %q: %d", token, status)
		}
	}
}

func TestSessionsSurviveRestartAndHideTokens(t *testing.T) {
	auth := bareAuth(t)
	tokens := auth.login(t)

	content, err := os.ReadFile(auth.stateFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), tokens.RefreshToken) {
		t.Fatal("refresh token stored in plain text")
	}
	info, err := os.Stat(auth.stateFile)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("mode: %v %v", info.Mode(), err)
	}

	// Новый экземпляр читает тот же файл
	repository, err := oauthRepository.NewOAuthRepository(auth.stateFile)
	if err != nil {
		t.Fatal(err)
	}
	service, err := oauthService.NewOAuthService(repository, oauthService.Config{
		PublicURL: auth.url, AccessTokenTTL: time.Minute, RefreshTokenTTL: time.Hour, PasswordHash: passwordHash(), ClientKey: []byte("client-key"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Refresh(t.Context(), tokens.RefreshToken); err != nil {
		t.Fatalf("refresh after restart: %v", err)
	}
}

func TestFormEscapesClientNameAndForbidsFraming(t *testing.T) {
	auth := bareAuth(t)
	body, _ := json.Marshal(oauthModel.RegisterClientReq{ClientName: "<script>alert(1)</script>", RedirectURIs: []string{testRedirect}})
	registered, err := http.Post(auth.url+"/oauth/register", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer registered.Body.Close()
	var res oauthModel.RegisterClientRes
	json.NewDecoder(registered.Body).Decode(&res)

	sum := sha256.Sum256([]byte(testVerifier))
	query := url.Values{"client_id": {res.ClientID}, "redirect_uri": {testRedirect}, "response_type": {"code"}, "code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"}}
	form, err := http.Get(auth.url + "/oauth/authorize?" + query.Encode())
	if err != nil {
		t.Fatal(err)
	}
	defer form.Body.Close()
	page, _ := io.ReadAll(form.Body)
	if strings.Contains(string(page), "<script>") {
		t.Error("client name is not escaped")
	}
	if !strings.Contains(string(page), "claude.ai") {
		t.Error("redirect host is not shown")
	}
	if form.Header.Get("X-Frame-Options") != "DENY" {
		t.Error("framing allowed")
	}
}
