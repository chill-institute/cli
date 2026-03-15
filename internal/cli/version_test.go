package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/chill-institute/cli/internal/buildinfo"
)

func TestVersionCommandOutputsBuildInfo(t *testing.T) {
	restore := currentBuildInfo
	currentBuildInfo = func() buildinfo.Info {
		return buildinfo.Info{
			Version:   "v1.2.3",
			Commit:    "abc1234",
			BuildDate: "2026-03-15T00:00:00Z",
		}
	}
	t.Cleanup(func() { currentBuildInfo = restore })

	stdout := &bytes.Buffer{}
	command := newVersionCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output["version"] != "v1.2.3" {
		t.Fatalf("version = %v", output["version"])
	}
	if output["commit"] != "abc1234" {
		t.Fatalf("commit = %v", output["commit"])
	}
}

func TestVersionCommandOutputsPrettyVersionLine(t *testing.T) {
	restore := currentBuildInfo
	currentBuildInfo = func() buildinfo.Info {
		return buildinfo.Info{
			Version:   "0.1.5",
			Commit:    "dacd5f16ad68251e65c87a0295a1992b12f00335",
			BuildDate: "2026-03-15T00:00:00Z",
		}
	}
	t.Cleanup(func() { currentBuildInfo = restore })

	stdout := &bytes.Buffer{}
	command := newVersionCommand(&appContext{
		opts:   &appOptions{output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "0.1.5 (dacd5f1)" {
		t.Fatalf("stdout = %q", got)
	}
}
