package cli

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chill-institute/cli/internal/buildinfo"
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

	app := &appContext{opts: &appOptions{profile: "staging"}}
	profile, err := app.activeProfile()
	if err != nil {
		t.Fatalf("activeProfile() error = %v", err)
	}
	if profile != "staging" {
		t.Fatalf("profile = %q, want %q", profile, "staging")
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
