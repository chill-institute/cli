package cli

import (
	"strings"
	"testing"
)

func TestReadLineTrimsInput(t *testing.T) {
	t.Parallel()

	stdout := &strings.Builder{}
	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(" token-123 \n"),
		stdout: stdout,
		stderr: &strings.Builder{},
	}

	line, err := app.readLine("token: ")
	if err != nil {
		t.Fatalf("readLine() error = %v", err)
	}
	if line != "token-123" {
		t.Fatalf("line = %q", line)
	}
	if stdout.String() != "token: " {
		t.Fatalf("prompt = %q", stdout.String())
	}
}

func TestOpenBrowserRejectsEmptyURL(t *testing.T) {
	t.Parallel()

	if err := openBrowser(" "); err == nil {
		t.Fatal("expected error")
	}
}
