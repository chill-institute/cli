package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRootCommandIncludesSchemaTopLevelCommand(t *testing.T) {
	command := NewRootCommand()
	expected := map[string]bool{
		"add-transfer":    true,
		"auth":            true,
		"list-top-movies": true,
		"schema":          true,
		"search":          true,
		"self-update":     true,
		"settings":        true,
		"user":            true,
		"version":         true,
		"whoami":          true,
	}

	for _, subcommand := range command.Commands() {
		if _, ok := expected[subcommand.Name()]; ok {
			expected[subcommand.Name()] = false
		}
	}

	for name, missing := range expected {
		if missing {
			t.Fatalf("missing top-level command %q", name)
		}
	}
}

func TestSchemaListsCommandsAndProceduresInJSON(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"schema", "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output struct {
		Commands   []schemaEntry `json:"commands"`
		Procedures []schemaEntry `json:"procedures"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(output.Commands) == 0 {
		t.Fatal("expected commands in schema output")
	}
	if len(output.Procedures) == 0 {
		t.Fatal("expected procedures in schema output")
	}
	if output.Commands[0].ID != "add-transfer" {
		t.Fatalf("first command id = %q, want %q", output.Commands[0].ID, "add-transfer")
	}
}

func TestSchemaCommandSearchReturnsMetadata(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"schema", "command", "search", "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output schemaEntry
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output.ID != "search" {
		t.Fatalf("ID = %q, want %q", output.ID, "search")
	}
	if output.AuthMode != "user" {
		t.Fatalf("AuthMode = %q, want %q", output.AuthMode, "user")
	}
	if output.LinkedProcedure != procedureUserSearch {
		t.Fatalf("LinkedProcedure = %q, want %q", output.LinkedProcedure, procedureUserSearch)
	}
	if output.Mutates {
		t.Fatal("search metadata unexpectedly marked mutating")
	}
	if output.SupportsDryRun {
		t.Fatal("search metadata unexpectedly supports dry run")
	}
	if !output.SupportsFields {
		t.Fatal("search metadata should support field selection")
	}
}

func TestSchemaProcedureSearchReturnsMetadata(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"schema", "procedure", procedureUserSearch, "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output schemaEntry
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output.Kind != "procedure" {
		t.Fatalf("Kind = %q, want %q", output.Kind, "procedure")
	}
	if output.AuthMode != "user" {
		t.Fatalf("AuthMode = %q, want %q", output.AuthMode, "user")
	}
}

func TestSearchDescribeOutputsMetadataWithoutExecuting(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"search", "--describe", "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output schemaEntry
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output.ID != "search" {
		t.Fatalf("ID = %q, want %q", output.ID, "search")
	}
}

func TestAuthLoginDescribeOutputsMetadataWithoutExecuting(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:    &appOptions{output: outputJSON},
		stdin:   strings.NewReader(""),
		stdout:  stdout,
		stderr:  &bytes.Buffer{},
		openURL: func(string) error { t.Fatal("openURL should not be called during describe"); return nil },
	})
	command.SetArgs([]string{"auth", "login", "--describe", "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output schemaEntry
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output.ID != "auth login" {
		t.Fatalf("ID = %q, want %q", output.ID, "auth login")
	}
}

func TestSchemaCommandAddTransferReturnsDryRunMetadata(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"schema", "command", "add-transfer", "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output schemaEntry
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !output.Mutates {
		t.Fatal("add-transfer metadata should be mutating")
	}
	if !output.SupportsDryRun {
		t.Fatal("add-transfer metadata should support dry run")
	}

	foundDryRun := false
	for _, input := range output.Inputs {
		if input.Name == "dry-run" {
			foundDryRun = true
			break
		}
	}
	if !foundDryRun {
		t.Fatal("add-transfer metadata missing dry-run input")
	}
}

func TestSchemaCommandWhoamiReturnsFieldSelectionMetadata(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"schema", "command", "whoami", "--output", "json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output schemaEntry
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !output.SupportsFields {
		t.Fatal("whoami metadata should support fields")
	}

	foundFields := false
	for _, input := range output.Inputs {
		if input.Name == "fields" {
			foundFields = true
			break
		}
	}
	if !foundFields {
		t.Fatal("whoami metadata missing fields input")
	}
}

func TestRootHelpStillWorks(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	command := newRootCommand(&appContext{
		opts:   &appOptions{output: outputPretty},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--help"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "schema") {
		t.Fatalf("help output missing schema command: %s", stdout.String())
	}
}
