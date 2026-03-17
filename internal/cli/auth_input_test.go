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

	token, skipOpen, localBrowser, skipVerify, err := resolveAuthLoginInput(app, "", `{"token":" token-123 ","no_browser":true,"local_browser":true,"skip_verify":true}`, false, false, false)
	if err != nil {
		t.Fatalf("resolveAuthLoginInput() error = %v", err)
	}
	if token != "token-123" {
		t.Fatalf("token = %q, want %q", token, "token-123")
	}
	if !skipOpen {
		t.Fatal("skipOpen = false, want true")
	}
	if !localBrowser {
		t.Fatal("localBrowser = false, want true")
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

	if _, _, _, _, err := resolveAuthLoginInput(app, "token-123", `{"token":"other"}`, false, false, false); err == nil {
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
		`{"local_browser":"yes"}`,
		`{"skip_verify":"yes"}`,
	}

	for _, rawRequest := range testCases {
		rawRequest := rawRequest
		t.Run(rawRequest, func(t *testing.T) {
			t.Parallel()

			if _, _, _, _, err := resolveAuthLoginInput(app, "", rawRequest, false, false, false); err == nil {
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

func TestResolveAuthLoginInputReturnsFlagValuesWithoutJSON(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	token, skipOpen, localBrowser, skipVerify, err := resolveAuthLoginInput(app, "token-123", "", true, true, true)
	if err != nil {
		t.Fatalf("resolveAuthLoginInput() error = %v", err)
	}
	if token != "token-123" || !skipOpen || !localBrowser || !skipVerify {
		t.Fatalf("resolved values = %q %v %v %v", token, skipOpen, localBrowser, skipVerify)
	}
}

func TestWebAuthTokenURLDerivesPublicHostFromAPIBaseURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		output string
	}{
		{name: "api subdomain", input: "https://api.chill.institute", output: "https://chill.institute/auth/cli-token"},
		{name: "non api host", input: "https://staging.chill.institute", output: "https://staging.chill.institute/auth/cli-token"},
		{name: "custom port", input: "http://api.localhost:3000", output: "http://localhost:3000/auth/cli-token"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := webAuthTokenURL(tc.input)
			if err != nil {
				t.Fatalf("webAuthTokenURL() error = %v", err)
			}
			if got != tc.output {
				t.Fatalf("webAuthTokenURL() = %q, want %q", got, tc.output)
			}
		})
	}
}

func TestWebAuthTokenURLRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"", "api.chill.institute"} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			if _, err := webAuthTokenURL(input); err == nil {
				t.Fatalf("webAuthTokenURL(%q) error = nil, want error", input)
			}
		})
	}
}
