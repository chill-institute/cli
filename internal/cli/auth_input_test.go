package cli

import (
	"strings"
	"testing"
)

func TestResolveAuthLoginInputFromJSON(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	token, skipOpen, skipVerify, err := resolveAuthLoginInput(app, "", `{"token":" token-123 ","no_browser":true,"skip_verify":true}`, false, false)
	if err != nil {
		t.Fatalf("resolveAuthLoginInput() error = %v", err)
	}
	if token != "token-123" {
		t.Fatalf("token = %q, want %q", token, "token-123")
	}
	if !skipOpen {
		t.Fatal("skipOpen = false, want true")
	}
	if !skipVerify {
		t.Fatal("skipVerify = false, want true")
	}
}

func TestResolveAuthLoginInputRejectsAmbiguousMixedInput(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	if _, _, _, err := resolveAuthLoginInput(app, "token-123", `{"token":"other"}`, false, false); err == nil {
		t.Fatal("resolveAuthLoginInput() error = nil, want ambiguity error")
	}
}

func TestResolveAuthLoginInputRejectsInvalidJSONTypes(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	testCases := []string{
		`{"token":true}`,
		`{"no_browser":"yes"}`,
		`{"skip_verify":"yes"}`,
	}

	for _, rawRequest := range testCases {
		rawRequest := rawRequest
		t.Run(rawRequest, func(t *testing.T) {
			t.Parallel()

			if _, _, _, err := resolveAuthLoginInput(app, "", rawRequest, false, false); err == nil {
				t.Fatalf("resolveAuthLoginInput(%s) error = nil, want error", rawRequest)
			}
		})
	}
}

func TestResolveAuthLogoutInputAcceptsExplicitTrue(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	request, err := resolveAuthLogoutInput(app, `{"clear_auth_token":true}`, true)
	if err != nil {
		t.Fatalf("resolveAuthLogoutInput() error = %v", err)
	}
	if request["clear_auth_token"] != true {
		t.Fatalf("request = %#v", request)
	}
	if request["had_auth_token"] != true {
		t.Fatalf("request = %#v", request)
	}
}

func TestResolveAuthLogoutInputRejectsInvalidJSONValue(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	for _, rawRequest := range []string{`{"clear_auth_token":false}`, `{"clear_auth_token":"yes"}`} {
		rawRequest := rawRequest
		t.Run(rawRequest, func(t *testing.T) {
			t.Parallel()

			if _, err := resolveAuthLogoutInput(app, rawRequest, false); err == nil {
				t.Fatalf("resolveAuthLogoutInput(%s) error = nil, want error", rawRequest)
			}
		})
	}
}
