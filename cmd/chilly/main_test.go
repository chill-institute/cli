package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{"--help"}, strings.NewReader(""), stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "chill.institute CLI for humans and agents") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestMainInvokesExitWithRunResult(t *testing.T) {
	originalExit := exit
	originalArgs := os.Args
	os.Args = []string{"chilly", "--help"}

	called := -1
	exit = func(code int) {
		called = code
	}
	t.Cleanup(func() {
		exit = originalExit
		os.Args = originalArgs
	})

	main()

	if called != 0 {
		t.Fatalf("exit code = %d", called)
	}
}
