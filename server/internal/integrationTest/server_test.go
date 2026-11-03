package integrationTest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	gitSyncEndpointMCP "server/internal/modules/gitSync/endpoint/mcp"
	gitSyncModel "server/internal/modules/gitSync/model"
	gitSyncRepository "server/internal/modules/gitSync/repository"
	gitSyncService "server/internal/modules/gitSync/service"
	noteEndpointMCP "server/internal/modules/note/endpoint/mcp"
	noteRepository "server/internal/modules/note/repository"
	noteService "server/internal/modules/note/service"
	mcpTransport "server/internal/utils/mcp"
	"strings"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// bearerTransport добавляет токен к запросам клиента
type bearerTransport struct {
	token string // Токен доступа
}

// RoundTrip выполняет запрос с токеном
func (b bearerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request = request.Clone(request.Context())
	request.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(request)
}

// setup поднимает сервер на временном клоне и подключает клиента
func setup(t *testing.T) (*httptest.Server, *sdk.ClientSession, string) {
	t.Helper()
	directory := t.TempDir()
	if output, err := exec.Command("git", "-C", directory, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("%s", output)
	}
	os.MkdirAll(filepath.Join(directory, "Здоровье"), 0o755)
	os.WriteFile(filepath.Join(directory, "Здоровье/x.md"), []byte("тайна"), 0o644)
	os.WriteFile(filepath.Join(directory, "a.md"), []byte("hello\n"), 0o644)

	notes, err := noteRepository.NewNoteRepository(directory, []string{"Здоровье"})
	if err != nil {
		t.Fatal(err)
	}
	gitSync := gitSyncService.NewGitSyncService(gitSyncRepository.NewGitSyncRepository(gitSyncModel.Config{
		Directory: directory, Remote: "origin", Branch: "main", AuthorName: "c", AuthorEmail: "c@c",
	}), time.Minute, time.Minute)
	server := mcpTransport.NewServer("test")
	noteEndpointMCP.Register(server, noteService.NewNoteService(notes, gitSync))
	gitSyncEndpointMCP.Register(server, gitSync)
	testServer := httptest.NewServer(mcpTransport.NewHandler(server, "secret"))
	t.Cleanup(testServer.Close)

	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	session, err := client.Connect(context.Background(), &sdk.StreamableClientTransport{
		Endpoint: testServer.URL + "/mcp", HTTPClient: &http.Client{Transport: bearerTransport{"secret"}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { session.Close() })
	return testServer, session, directory
}

// call вызывает инструмент и разбирает структурированный ответ
func call(t *testing.T, session *sdk.ClientSession, name string, arguments map[string]any) (*sdk.CallToolResult, map[string]any) {
	t.Helper()
	result, err := session.CallTool(context.Background(), &sdk.CallToolParams{Name: name, Arguments: arguments})
	if err != nil {
		t.Fatal(err)
	}
	var output map[string]any
	if encoded, err := json.Marshal(result.StructuredContent); err == nil {
		json.Unmarshal(encoded, &output)
	}
	return result, output
}

func TestRejectsMissingOrWrongToken(t *testing.T) {
	testServer, _, _ := setup(t)
	for _, header := range []string{"", "Bearer wrong", "secret"} {
		request, _ := http.NewRequest(http.MethodPost, testServer.URL+"/mcp", strings.NewReader("{}"))
		if header != "" {
			request.Header.Set("Authorization", header)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusUnauthorized {
			t.Errorf("header %q: status %d", header, response.StatusCode)
		}
	}
}

func TestToolSetHasNoDelete(t *testing.T) {
	_, session, _ := setup(t)
	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, tool := range tools.Tools {
		found[tool.Name] = true
	}
	for _, name := range []string{"list", "read", "search", "write", "replace", "append", "move", "commit"} {
		if !found[name] {
			t.Errorf("missing tool %s", name)
		}
	}
	if len(found) != 8 {
		t.Errorf("unexpected tools: %v", found)
	}
}

func TestReadEditCommitFlow(t *testing.T) {
	_, session, directory := setup(t)
	_, page := call(t, session, "read", map[string]any{"path": "a.md"})
	result, _ := call(t, session, "replace", map[string]any{"path": "a.md", "old_text": "hello", "new_text": "bye", "expected_hash": page["hash"]})
	if result.IsError {
		t.Fatalf("replace: %+v", result.Content)
	}
	if result, _ := call(t, session, "replace", map[string]any{"path": "a.md", "old_text": "bye", "new_text": "x", "expected_hash": page["hash"]}); !result.IsError {
		t.Fatal("stale hash must fail")
	}
	_, output := call(t, session, "commit", map[string]any{"message": "edit"})
	if output["committed"] != true {
		t.Fatalf("commit: %v", output)
	}
	content, _ := os.ReadFile(filepath.Join(directory, "a.md"))
	if string(content) != "bye\n" {
		t.Fatalf("content %q", content)
	}
}

func TestDeniedFolderIsInvisible(t *testing.T) {
	_, session, _ := setup(t)
	if result, _ := call(t, session, "read", map[string]any{"path": "Здоровье/x.md"}); !result.IsError {
		t.Fatal("read must fail")
	}
	_, output := call(t, session, "search", map[string]any{"query": "тайна"})
	if hits, _ := output["hits"].([]any); len(hits) != 0 {
		t.Fatalf("leak: %v", hits)
	}
}

func TestInternalErrorDetailsAreHidden(t *testing.T) {
	_, session, directory := setup(t)
	os.RemoveAll(filepath.Join(directory, ".git"))
	os.WriteFile(filepath.Join(directory, "b.md"), []byte("x"), 0o644)
	result, _ := call(t, session, "commit", map[string]any{"message": "m"})
	if !result.IsError {
		t.Fatal("commit must fail")
	}
	encoded, _ := json.Marshal(result.Content)
	if strings.Contains(string(encoded), directory) || strings.Contains(string(encoded), "git") {
		t.Fatalf("details leaked: %s", encoded)
	}
}
