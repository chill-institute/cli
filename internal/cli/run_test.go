package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chill-institute/cli/internal/config"
)

func TestRunReturnsUsageExitCodeAndJSONErrorEnvelope(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"search", "--output", "json"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeUsage) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeUsage)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stderr) error = %v", err)
	}
	if output["kind"] != "usage" {
		t.Fatalf("kind = %v, want %q", output["kind"], "usage")
	}
	if output["code"] != "missing_query" {
		t.Fatalf("code = %v, want %q", output["code"], "missing_query")
	}
}

func TestRunReturnsAuthExitCodeAndJSONErrorEnvelope(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"--config", configPath, "whoami", "--output", "json"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeAuth) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeAuth)
	}

	var output map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stderr) error = %v", err)
	}
	if output["kind"] != "auth" {
		t.Fatalf("kind = %v, want %q", output["kind"], "auth")
	}
	if output["code"] != "missing_auth_token" {
		t.Fatalf("code = %v, want %q", output["code"], "missing_auth_token")
	}
}

func TestRunReturnsAPIExitCodeAndJSONErrorEnvelope(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(`{"code":"search_failed","message":"search backend failed","request_id":"req-500"}`))
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
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"--config", configPath, "search", "--query", "dune", "--output", "json"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeAPI) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeAPI)
	}

	var output map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stderr) error = %v", err)
	}
	if output["kind"] != "api" {
		t.Fatalf("kind = %v, want %q", output["kind"], "api")
	}
	if output["code"] != "search_failed" {
		t.Fatalf("code = %v, want %q", output["code"], "search_failed")
	}
	if output["request_id"] != "req-500" {
		t.Fatalf("request_id = %v, want %q", output["request_id"], "req-500")
	}
	if output["status_code"] != float64(http.StatusInternalServerError) {
		t.Fatalf("status_code = %v, want %d", output["status_code"], http.StatusInternalServerError)
	}
}

func TestRunReturnsInternalExitCodeAndJSONErrorEnvelope(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte("{invalid"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"--config", configPath, "whoami", "--output", "json"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeInternal) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeInternal)
	}

	var output map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stderr) error = %v", err)
	}
	if output["kind"] != "internal" {
		t.Fatalf("kind = %v, want %q", output["kind"], "internal")
	}
	if output["code"] != "config_load_failed" {
		t.Fatalf("code = %v, want %q", output["code"], "config_load_failed")
	}
}
