package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chill-institute/cli/internal/config"
)

func TestAuthLoginWithTokenSavesConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v4/chill.v4.UserService/GetUserProfile" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q", got)
		}
		_, _ = writer.Write([]byte(`{"user_id":"user-123"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app := &appContext{
		opts:    &appOptions{configPath: configPath, apiURL: server.URL, output: outputJSON},
		stdin:   strings.NewReader(""),
		stdout:  stdout,
		stderr:  stderr,
		openURL: func(string) error { return nil },
	}

	command := newAuthLoginCommand(app)
	command.SetArgs([]string{"--token", "test-token"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthToken != "test-token" {
		t.Fatalf("AuthToken = %q, want %q", cfg.AuthToken, "test-token")
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["status"] != "ok" {
		t.Fatalf("status = %v", output["status"])
	}
	if output["saved"] != true {
		t.Fatalf("saved = %v", output["saved"])
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestAuthLogoutClearsToken(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: "http://localhost:8080", AuthToken: "token-1"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	app := &appContext{
		opts:    &appOptions{configPath: configPath, output: outputJSON},
		stdin:   strings.NewReader(""),
		stdout:  stdout,
		stderr:  &bytes.Buffer{},
		openURL: func(string) error { return nil },
	}

	command := newAuthLogoutCommand(app)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthToken != "" {
		t.Fatalf("AuthToken = %q, want empty", cfg.AuthToken)
	}
}

func TestAuthLogoutDryRunKeepsToken(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: "https://api.binge.institute", AuthToken: "token-1"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	app := &appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	}

	command := newAuthLogoutCommand(app)
	command.SetArgs([]string{"--dry-run"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthToken != "token-1" {
		t.Fatalf("AuthToken = %q, want %q", cfg.AuthToken, "token-1")
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["dry_run"] != true {
		t.Fatalf("dry_run = %v, want true", output["dry_run"])
	}
}

func TestSearchCommandUsesStoredToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v4/chill.v4.UserService/Search" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer saved-token" {
			t.Fatalf("Authorization = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["query"] != "blade runner" {
			t.Fatalf("query = %v", payload["query"])
		}
		_, _ = writer.Write([]byte(`{"query":"blade runner","results":[]}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newSearchCommand(&appContext{
		opts:    &appOptions{configPath: configPath, output: outputJSON},
		stdin:   strings.NewReader(""),
		stdout:  stdout,
		stderr:  &bytes.Buffer{},
		openURL: func(string) error { return nil },
	})
	command.SetArgs([]string{"--query", "blade runner"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["query"] != "blade runner" {
		t.Fatalf("output query = %v", output["query"])
	}
}

func TestSearchCommandFieldsFiltersResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"results":[{"title":"Dune","magnetLink":"magnet:?xt=urn:btih:dune","size":"1.4 GB"}],"request_id":"req-123"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newSearchCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--query", "dune", "--fields", "results.title,request_id"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if _, ok := output["request_id"]; !ok {
		t.Fatal("expected request_id in filtered output")
	}
	results, ok := output["results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("results = %#v", output["results"])
	}
	first, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("first result = %#v", results[0])
	}
	if first["title"] != "Dune" {
		t.Fatalf("first.title = %v", first["title"])
	}
	if _, ok := first["magnetLink"]; ok {
		t.Fatalf("unexpected magnetLink in %#v", first)
	}
}

func TestSearchCommandPrettyOutputShowsReadableSummary(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"query":"dune","results":[{"title":"Dune","magnetLink":"magnet:?xt=urn:btih:dune","size":"1.4 GB","seeders":120}]}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newSearchCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--query", "dune"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	rendered := stdout.String()
	for _, expected := range []string{"Results: 1", "Query: dune", "1. Dune", "Size: 1.4 GB", "Magnet: magnet:?xt=urn:btih:dune"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("pretty output missing %q in %q", expected, rendered)
		}
	}
}

func TestSearchCommandPrettyOutputTruncatesLongLists(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"query":"dune","results":[{"title":"One"},{"title":"Two"},{"title":"Three"},{"title":"Four"},{"title":"Five"},{"title":"Six"},{"title":"Seven"},{"title":"Eight"},{"title":"Nine"},{"title":"Ten"},{"title":"Eleven"}]}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newSearchCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--query", "dune"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	rendered := stdout.String()
	if strings.Contains(rendered, "11. Eleven") {
		t.Fatalf("pretty output should truncate long lists: %q", rendered)
	}
	if !strings.Contains(rendered, "... 1 more results omitted.") {
		t.Fatalf("pretty output missing truncation notice: %q", rendered)
	}
}

func TestSettingsSetAndGetAPIBaseURL(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	setOutput := &bytes.Buffer{}
	app := &appContext{
		opts:    &appOptions{configPath: configPath, output: outputJSON},
		stdin:   strings.NewReader(""),
		stdout:  setOutput,
		stderr:  &bytes.Buffer{},
		openURL: func(string) error { return nil },
	}
	setCommand := newSettingsCommand(app)
	setCommand.SetArgs([]string{"set", "api-base-url", "https://api.chill.test"})
	if err := setCommand.Execute(); err != nil {
		t.Fatalf("set Execute() error = %v", err)
	}

	getOutput := &bytes.Buffer{}
	app.stdout = getOutput
	getCommand := newSettingsCommand(app)
	getCommand.SetArgs([]string{"get", "api-base-url"})
	if err := getCommand.Execute(); err != nil {
		t.Fatalf("get Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(getOutput.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["value"] != "https://api.chill.test" {
		t.Fatalf("value = %v, want %q", output["value"], "https://api.chill.test")
	}
}

func TestSettingsSetDryRunDoesNotSaveAPIBaseURL(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Default()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	app := &appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	}

	command := newSettingsCommand(app)
	command.SetArgs([]string{"set", "api-base-url", "https://api.chill.test", "--dry-run"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIBaseURL != config.Default().APIBaseURL {
		t.Fatalf("APIBaseURL = %q, want %q", cfg.APIBaseURL, config.Default().APIBaseURL)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["command"] != "settings set" {
		t.Fatalf("command = %v, want %q", output["command"], "settings set")
	}
}

func TestSettingsShowRedactsToken(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: "https://api.chill.test", AuthToken: "secret-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newSettingsCommand(&appContext{
		opts:    &appOptions{configPath: configPath, output: outputJSON},
		stdin:   strings.NewReader(""),
		stdout:  stdout,
		stderr:  &bytes.Buffer{},
		openURL: func(string) error { return nil },
	})
	command.SetArgs([]string{"show"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["auth_token"] != redactedToken {
		t.Fatalf("auth_token = %v, want %q", output["auth_token"], redactedToken)
	}
}

func TestListTopMoviesUsesStoredToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v4/chill.v4.UserService/GetTopMovies" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer saved-token" {
			t.Fatalf("Authorization = %q", got)
		}
		_, _ = writer.Write([]byte(`{"movies":[{"title":"Dune"}]}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newListTopMoviesCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), `"title":"Dune"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestListTopMoviesFieldsFiltersResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"movies":[{"title":"Dune","year":2021}]}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newListTopMoviesCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--fields", "movies.title"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if strings.Contains(stdout.String(), `"year"`) {
		t.Fatalf("stdout should not include year: %q", stdout.String())
	}
}

func TestListTopMoviesPrettyOutputShowsReadableSummary(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"movies":[{"title":"Dune","year":2021},{"title":"Arrival","year":2016}]}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newListTopMoviesCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	rendered := stdout.String()
	for _, expected := range []string{"Movies: 2", "1. Dune (2021)", "2. Arrival (2016)"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("pretty output missing %q in %q", expected, rendered)
		}
	}
}

func TestSettingsPathOutputsResolvedStorePath(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	command := newSettingsCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"path"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output["path"] != configPath {
		t.Fatalf("path = %v, want %q", output["path"], configPath)
	}
}

func TestWhoamiFieldsFiltersResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"username":"dunefan","email":"dune@example.com","userId":"7"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newWhoamiCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--fields", "username,email"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["username"] != "dunefan" || output["email"] != "dune@example.com" {
		t.Fatalf("output = %#v", output)
	}
	if _, ok := output["userId"]; ok {
		t.Fatalf("unexpected userId in %#v", output)
	}
}

func TestWhoamiPrettyOutputShowsReadableSummary(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"username":"dunefan","email":"dune@example.com","userId":"7"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newWhoamiCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	rendered := stdout.String()
	for _, expected := range []string{"Username: dunefan", "Email: dune@example.com", "User ID: 7"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("pretty output missing %q in %q", expected, rendered)
		}
	}
}

func TestUserSettingsGetFieldsFiltersResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v4/chill.v4.UserService/GetUserSettings" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"showTopMovies":true,"sortBy":"seeders","sortDirection":"desc"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newUserCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"settings", "get", "--fields", "showTopMovies,sortBy"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["showTopMovies"] != true || output["sortBy"] != "seeders" {
		t.Fatalf("output = %#v", output)
	}
	if _, ok := output["sortDirection"]; ok {
		t.Fatalf("unexpected sortDirection in %#v", output)
	}
}

func TestUserProfilePrettyOutputShowsReadableSummary(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v4/chill.v4.UserService/GetUserProfile" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		_, _ = writer.Write([]byte(`{"username":"dunefan","email":"dune@example.com","userId":"7"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: server.URL, AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	command := newUserCommand(&appContext{
		opts:   &appOptions{configPath: configPath, output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"profile"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "Username: dunefan") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
