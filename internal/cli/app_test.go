package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chill-institute/chill-cli/internal/buildinfo"
)

func TestReadLineTrimsInput(t *testing.T) {
	t.Parallel()

	stderr := &strings.Builder{}
	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(" token-123 \n"),
		stdout: &strings.Builder{},
		stderr: stderr,
	}

	line, err := app.readLine("token: ")
	if err != nil {
		t.Fatalf("readLine() error = %v", err)
	}
	if line != "token-123" {
		t.Fatalf("line = %q", line)
	}
	if stderr.String() != "token: " {
		t.Fatalf("prompt = %q", stderr.String())
	}
}

func TestOpenBrowserRejectsEmptyURL(t *testing.T) {
	t.Parallel()

	if err := openBrowser(" "); err == nil {
		t.Fatal("expected error")
	}
}

func TestActiveProfileUsesDevDefaultForDevBuilds(t *testing.T) {
	t.Parallel()

	original := currentBuildInfo
	currentBuildInfo = func() buildinfo.Info {
		return buildinfo.Info{Version: "dev", Commit: "test", BuildDate: "test"}
	}
	t.Cleanup(func() { currentBuildInfo = original })

	app := &appContext{opts: &appOptions{}}
	profile, err := app.activeProfile()
	if err != nil {
		t.Fatalf("activeProfile() error = %v", err)
	}
	if profile != "dev" {
		t.Fatalf("profile = %q, want %q", profile, "dev")
	}
}

func TestActiveProfileUsesExplicitProfile(t *testing.T) {
	t.Parallel()

	app := &appContext{opts: &appOptions{profile: "production"}}
	profile, err := app.activeProfile()
	if err != nil {
		t.Fatalf("activeProfile() error = %v", err)
	}
	if profile != "production" {
		t.Fatalf("profile = %q, want %q", profile, "production")
	}
}

func TestActiveProfileRejectsInvalidProfile(t *testing.T) {
	t.Parallel()

	app := &appContext{opts: &appOptions{profile: "../prod"}}
	profile, err := app.activeProfile()
	if err == nil {
		t.Fatalf("activeProfile() error = nil, profile = %q", profile)
	}
	if !strings.Contains(err.Error(), "invalid profile") {
		t.Fatalf("error = %v", err)
	}
}

func TestShouldShowProgressOnlyForPrettyTerminalOutput(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:          &appOptions{output: outputPretty},
		stderr:        &bytes.Buffer{},
		isTerminal:    func(io.Writer) bool { return true },
		progressEvery: time.Millisecond,
	}
	if !app.shouldShowProgress() {
		t.Fatal("expected pretty terminal output to enable progress")
	}

	app.opts.output = outputJSON
	if app.shouldShowProgress() {
		t.Fatal("did not expect JSON output to enable progress")
	}

	app.opts.output = outputPretty
	app.isTerminal = func(io.Writer) bool { return false }
	if app.shouldShowProgress() {
		t.Fatal("did not expect non-terminal output to enable progress")
	}
}

func TestReadJSONFlagSupportsStdin(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader("{\"url\":\"magnet:?xt=urn:btih:test\"}\n"),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	payload, err := app.decodeJSONObjectFlag("@-", "--json")
	if err != nil {
		t.Fatalf("decodeJSONObjectFlag() error = %v", err)
	}
	if payload["url"] != "magnet:?xt=urn:btih:test" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestRunDefaultsToJSONWhenStdoutIsNotATerminal(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"version"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal(stdout) error = %v; stdout=%q", err, stdout.String())
	}
	if output["name"] != "chilly" {
		t.Fatalf("output = %#v", output)
	}
}

func TestRunHonorsExplicitPrettyOutputWhenStdoutIsNotATerminal(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"version", "--output", "pretty"}, strings.NewReader(""), stdout, stderr)
	if exitCode != int(exitCodeSuccess) {
		t.Fatalf("exitCode = %d, want %d", exitCode, exitCodeSuccess)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if strings.Contains(stdout.String(), `"name"`) {
		t.Fatalf("stdout = %q, want pretty output", stdout.String())
	}
}

func TestWithProgressWritesToStderrAndClears(t *testing.T) {
	t.Parallel()

	stderr := &recordingWriter{wrote: make(chan struct{}, 4)}
	app := &appContext{
		opts:          &appOptions{output: outputPretty},
		stderr:        stderr,
		isTerminal:    func(io.Writer) bool { return true },
		newTicker:     func(time.Duration) progressTicker { return fakeProgressTicker{channel: make(chan time.Time)} },
		progressLabel: "Loading",
		progressEvery: time.Millisecond,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.withProgress(func() error {
			<-stderr.wrote
			return nil
		})
	}()
	if err := <-errCh; err != nil {
		t.Fatalf("withProgress() error = %v", err)
	}

	rendered := stderr.String()
	if !strings.Contains(rendered, "Loading") {
		t.Fatalf("stderr = %q, want loading indicator", rendered)
	}
	if !strings.Contains(rendered, "\r") {
		t.Fatalf("stderr = %q, want carriage-return progress output", rendered)
	}
}

func TestNormalizeJSONAndWriters(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeJSON([]byte(" "), outputJSON, nil)
	if err != nil {
		t.Fatalf("normalizeJSON(empty) error = %v", err)
	}
	if string(normalized) != "{}" {
		t.Fatalf("normalizeJSON(empty) = %q", normalized)
	}

	if _, err := normalizeJSON([]byte("{"), outputJSON, nil); err == nil {
		t.Fatal("normalizeJSON(invalid) error = nil, want error")
	}

	stdout := &bytes.Buffer{}
	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdout: stdout,
		stderr: &bytes.Buffer{},
	}
	if err := app.writeSelectedResponseBody([]byte(`{"user":{"name":"sample-user"}}`), mustFieldSelection(t, "user.name")); err != nil {
		t.Fatalf("writeSelectedResponseBody() error = %v", err)
	}
	if strings.TrimSpace(stdout.String()) != `{"user":{"name":"sample-user"}}` {
		t.Fatalf("stdout = %q", stdout.String())
	}

	stdout.Reset()
	if err := app.writeJSONPayload(map[string]any{"status": "ok"}); err != nil {
		t.Fatalf("writeJSONPayload() error = %v", err)
	}
	if strings.TrimSpace(stdout.String()) != `{"status":"ok"}` {
		t.Fatalf("stdout = %q", stdout.String())
	}

	app.opts.output = outputPretty
	stdout.Reset()
	if err := app.writeAnyWithRenderer(map[string]any{"name": "sample-user"}, nil, func(value any) (string, bool, error) {
		return "Name: sample-user", true, nil
	}); err != nil {
		t.Fatalf("writeAnyWithRenderer() error = %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "Name: sample-user" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestNormalizeJSONSupportsNDJSONArrays(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeJSON([]byte(`[{"name":"one"},{"name":"two"}]`), outputNDJSON, nil)
	if err != nil {
		t.Fatalf("normalizeJSON(ndjson array) error = %v", err)
	}
	lines := strings.Split(string(normalized), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %#v, want 2 lines", lines)
	}
	if lines[0] != `{"name":"one"}` || lines[1] != `{"name":"two"}` {
		t.Fatalf("ndjson = %q", normalized)
	}
}

func TestNormalizeJSONSupportsNDJSONResponseCollections(t *testing.T) {
	t.Parallel()

	normalized, err := normalizeJSON([]byte(`{"query":"dune","results":[{"title":"Dune"},{"title":"Dune Part Two"}]}`), outputNDJSON, nil)
	if err != nil {
		t.Fatalf("normalizeJSON(ndjson object) error = %v", err)
	}
	lines := strings.Split(string(normalized), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %#v, want 2 lines", lines)
	}

	var first map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("json.Unmarshal(first) error = %v", err)
	}
	if first["path"] != "results" || first["index"] != float64(0) {
		t.Fatalf("first = %#v", first)
	}
	item, ok := first["item"].(map[string]any)
	if !ok || item["title"] != "Dune" {
		t.Fatalf("item = %#v", first["item"])
	}
	context, ok := first["context"].(map[string]any)
	if !ok || context["query"] != "dune" {
		t.Fatalf("context = %#v", first["context"])
	}
}

func TestWriteSelectedResponseBodyWithRendererFallbackAndErrors(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	app := &appContext{
		opts:   &appOptions{output: outputPretty},
		stdout: stdout,
		stderr: &bytes.Buffer{},
	}

	if err := app.writeSelectedResponseBodyWithRenderer([]byte(`{"status":"ok"}`), nil, func(value any) (string, bool, error) {
		return "", false, nil
	}); err != nil {
		t.Fatalf("writeSelectedResponseBodyWithRenderer(fallback) error = %v", err)
	}
	if !strings.Contains(stdout.String(), `"status": "ok"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}

	if err := app.writeSelectedResponseBodyWithRenderer([]byte(`{"status":"ok"}`), nil, func(value any) (string, bool, error) {
		return "", false, errors.New("boom")
	}); err == nil {
		t.Fatal("writeSelectedResponseBodyWithRenderer(error) error = nil, want error")
	}
}

func TestWithProgressNoopPaths(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:       &appOptions{output: outputJSON},
		stderr:     &bytes.Buffer{},
		isTerminal: func(io.Writer) bool { return true },
	}
	if err := app.withProgress(nil); err != nil {
		t.Fatalf("withProgress(nil) error = %v", err)
	}

	called := false
	if err := app.withProgress(func() error {
		called = true
		return nil
	}); err != nil {
		t.Fatalf("withProgress(no-progress) error = %v", err)
	}
	if !called {
		t.Fatal("withProgress(no-progress) did not call function")
	}
}

func mustFieldSelection(t *testing.T, value string) *fieldSelection {
	t.Helper()

	selection, err := parseFieldSelection(value)
	if err != nil {
		t.Fatalf("parseFieldSelection() error = %v", err)
	}
	return selection
}

type fakeProgressTicker struct {
	channel <-chan time.Time
}

func (ticker fakeProgressTicker) C() <-chan time.Time {
	return ticker.channel
}

func (ticker fakeProgressTicker) Stop() {}

type recordingWriter struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	wrote  chan struct{}
}

func (writer *recordingWriter) Write(payload []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.wrote != nil {
		select {
		case writer.wrote <- struct{}{}:
		default:
		}
	}
	return writer.buffer.Write(payload)
}

func (writer *recordingWriter) String() string {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.buffer.String()
}
