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

	"github.com/chill-institute/chill-institute-cli/internal/buildinfo"
	"github.com/chill-institute/chill-institute-cli/internal/config"
)

func TestRunAddTransferDryRunSkipsAuthAndReturnsPreview(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{
		"--config", configPath,
		"add-transfer",
		"--url", "magnet:?xt=urn:btih:dryrun",
		"--dry-run",
		"--output", "json",
	}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v", err)
	}
	if output["dry_run"] != true {
		t.Fatalf("dry_run = %v, want true", output["dry_run"])
	}
	if output["command"] != "add-transfer" {
		t.Fatalf("command = %v, want %q", output["command"], "add-transfer")
	}
	if output["procedure"] != procedureUserAddTransfer {
		t.Fatalf("procedure = %v, want %q", output["procedure"], procedureUserAddTransfer)
	}

	request, ok := output["request"].(map[string]any)
	if !ok {
		t.Fatalf("request = %#v, want object", output["request"])
	}
	if request["url"] != "magnet:?xt=urn:btih:dryrun" {
		t.Fatalf("request.url = %v, want %q", request["url"], "magnet:?xt=urn:btih:dryrun")
	}
}

func TestRunUserSettingsSetDryRunSkipsAuthAndReturnsPreview(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{
		"--config", configPath,
		"user", "settings", "set",
		"--json", `{"showTopMovies":true}`,
		"--dry-run",
		"--output", "json",
	}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v", err)
	}
	if output["dry_run"] != true {
		t.Fatalf("dry_run = %v, want true", output["dry_run"])
	}
	if output["command"] != "user settings set" {
		t.Fatalf("command = %v, want %q", output["command"], "user settings set")
	}
	if output["procedure"] != procedureUserSaveUserSettings {
		t.Fatalf("procedure = %v, want %q", output["procedure"], procedureUserSaveUserSettings)
	}

	request, ok := output["request"].(map[string]any)
	if !ok {
		t.Fatalf("request = %#v, want object", output["request"])
	}
	settings, ok := request["settings"].(map[string]any)
	if !ok {
		t.Fatalf("request.settings = %#v, want object", request["settings"])
	}
	if settings["showTopMovies"] != true {
		t.Fatalf("request.settings.showTopMovies = %v, want true", settings["showTopMovies"])
	}
}

func TestRunUserSettingsSetPatchDryRunSkipsAuthAndReturnsPreview(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{
		"--config", configPath,
		"user", "settings", "set",
		"show-top-movies", "true",
		"--dry-run",
		"--output", "json",
	}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v", err)
	}
	patch, ok := output["request"].(map[string]any)["patch"].(map[string]any)
	if !ok {
		t.Fatalf("request.patch = %#v, want object", output["request"])
	}
	if patch["field"] != "showTopMovies" {
		t.Fatalf("patch.field = %v, want %q", patch["field"], "showTopMovies")
	}
	if patch["value"] != true {
		t.Fatalf("patch.value = %v, want true", patch["value"])
	}
}

func TestRunSettingsSetDryRunReturnsLocalPreview(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{
		"--config", configPath,
		"settings", "set",
		"api-base-url", "https://api.chill.test",
		"--dry-run",
		"--output", "json",
	}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v", err)
	}
	if output["dry_run"] != true {
		t.Fatalf("dry_run = %v, want true", output["dry_run"])
	}
	if output["command"] != "settings set" {
		t.Fatalf("command = %v, want %q", output["command"], "settings set")
	}
	if _, ok := output["procedure"]; ok {
		t.Fatalf("procedure = %#v, want omitted for local preview", output["procedure"])
	}

	request, ok := output["request"].(map[string]any)
	if !ok {
		t.Fatalf("request = %#v, want object", output["request"])
	}
	if request["key"] != "api-base-url" {
		t.Fatalf("request.key = %v, want %q", request["key"], "api-base-url")
	}
	if request["value"] != "https://api.chill.test" {
		t.Fatalf("request.value = %v, want %q", request["value"], "https://api.chill.test")
	}
}

func TestRunAuthLogoutDryRunReturnsLocalPreview(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	if err := store.Save(config.Config{APIBaseURL: "https://api.binge.institute", AuthToken: "saved-token"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{
		"--config", configPath,
		"auth", "logout",
		"--dry-run",
		"--output", "json",
	}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v", err)
	}
	if output["command"] != "auth logout" {
		t.Fatalf("command = %v, want %q", output["command"], "auth logout")
	}
	request, ok := output["request"].(map[string]any)
	if !ok {
		t.Fatalf("request = %#v, want object", output["request"])
	}
	if request["clear_auth_token"] != true {
		t.Fatalf("request.clear_auth_token = %v, want true", request["clear_auth_token"])
	}
	if request["had_auth_token"] != true {
		t.Fatalf("request.had_auth_token = %v, want true", request["had_auth_token"])
	}
}

func TestRunSettingsPathUsesDevProfileByDefaultForDevBuilds(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	original := currentBuildInfo
	currentBuildInfo = func() buildinfo.Info {
		return buildinfo.Info{Version: "dev", Commit: "test", BuildDate: "test"}
	}
	t.Cleanup(func() { currentBuildInfo = original })

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"settings", "path", "--output", "json"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v", err)
	}

	wantPath := filepath.Join(configHome, "chilly", "profiles", "dev", "config.json")
	if output["path"] != wantPath {
		t.Fatalf("path = %v, want %q", output["path"], wantPath)
	}
	if output["profile"] != "dev" {
		t.Fatalf("profile = %v, want %q", output["profile"], "dev")
	}
}

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
