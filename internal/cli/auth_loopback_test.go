package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/chill-institute/cli/internal/config"
)

func TestAuthLoginBrowserFlowCapturesLoopbackToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v4/chill.v4.UserService/GetUserProfile" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer loopback-token" {
			t.Fatalf("Authorization = %q", got)
		}
		_, _ = writer.Write([]byte(`{"user_id":"user-123"}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.json")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	app := &appContext{
		opts:            &appOptions{configPath: configPath, apiURL: server.URL, output: outputJSON},
		stdin:           strings.NewReader(""),
		stdout:          stdout,
		stderr:          stderr,
		authFlowTimeout: 2 * time.Second,
		openURL: func(rawURL string) error {
			loginURL, err := url.Parse(rawURL)
			if err != nil {
				t.Fatalf("url.Parse(loginURL) error = %v", err)
			}
			if loginURL.Path != "/auth/putio/start" {
				t.Fatalf("login path = %q", loginURL.Path)
			}

			successURL := strings.TrimSpace(loginURL.Query().Get("success_url"))
			if successURL == "" {
				t.Fatal("missing success_url")
			}

			callbackResponse, err := http.Get(successURL)
			if err != nil {
				t.Fatalf("http.Get(successURL) error = %v", err)
			}
			defer func() {
				_ = callbackResponse.Body.Close()
			}()

			if callbackResponse.StatusCode != http.StatusOK {
				t.Fatalf("callback status = %d, want %d", callbackResponse.StatusCode, http.StatusOK)
			}

			body, err := io.ReadAll(callbackResponse.Body)
			if err != nil {
				t.Fatalf("io.ReadAll(callbackResponse.Body) error = %v", err)
			}

			match := regexp.MustCompile(`data-token-endpoint="([^"]+)"`).FindStringSubmatch(string(body))
			if len(match) != 2 {
				t.Fatalf("token endpoint not found in callback page: %s", string(body))
			}

			postBody := strings.NewReader(`{"auth_token":"loopback-token"}`)
			postResponse, err := http.Post(match[1], "application/json", postBody)
			if err != nil {
				t.Fatalf("http.Post(tokenEndpoint) error = %v", err)
			}
			defer func() {
				_ = postResponse.Body.Close()
			}()

			if postResponse.StatusCode != http.StatusOK {
				t.Fatalf("token post status = %d, want %d", postResponse.StatusCode, http.StatusOK)
			}

			return nil
		},
	}

	command := newAuthLoginCommand(app)
	command.SetArgs(nil)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthToken != "loopback-token" {
		t.Fatalf("AuthToken = %q, want %q", cfg.AuthToken, "loopback-token")
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("output json decode error: %v", err)
	}
	if output["status"] != "ok" {
		t.Fatalf("status = %v", output["status"])
	}
	if output["saved"] != true {
		t.Fatalf("saved = %v", output["saved"])
	}
	if !strings.Contains(stderr.String(), "Open this URL to authenticate:") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
