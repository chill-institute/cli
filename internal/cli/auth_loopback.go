package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type loopbackAuthFlow struct {
	baseURL      string
	loginURL     string
	callbackPath string
	tokenPath    string
	listener     net.Listener
	server       *http.Server
	tokenCh      chan string
}

func newLoopbackAuthFlow(apiBaseURL string) (*loopbackAuthFlow, error) {
	nonce, err := newLoopbackAuthNonce()
	if err != nil {
		return nil, fmt.Errorf("create oauth callback nonce: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen for oauth callback: %w", err)
	}

	baseURL := "http://" + listener.Addr().String()
	callbackPath := "/auth/callback/" + nonce
	tokenPath := "/auth/token/" + nonce

	successURL, err := url.JoinPath(baseURL, callbackPath)
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("build oauth callback url: %w", err)
	}

	loginURL, err := url.JoinPath(strings.TrimRight(strings.TrimSpace(apiBaseURL), "/"), "/auth/putio/start")
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("build oauth start url: %w", err)
	}

	parsedLoginURL, err := url.Parse(loginURL)
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("parse oauth start url: %w", err)
	}
	query := parsedLoginURL.Query()
	query.Set("success_url", successURL)
	parsedLoginURL.RawQuery = query.Encode()

	flow := &loopbackAuthFlow{
		baseURL:      baseURL,
		loginURL:     parsedLoginURL.String(),
		callbackPath: callbackPath,
		tokenPath:    tokenPath,
		listener:     listener,
		tokenCh:      make(chan string, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(callbackPath, flow.handleCallback)
	mux.HandleFunc(tokenPath, flow.handleToken)
	flow.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return flow, nil
}

func (flow *loopbackAuthFlow) start(errCh chan<- error) {
	go func() {
		if err := flow.server.Serve(flow.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
}

func (flow *loopbackAuthFlow) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := flow.server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (flow *loopbackAuthFlow) waitForToken(ctx context.Context, errCh <-chan error) (string, error) {
	select {
	case token := <-flow.tokenCh:
		return token, nil
	case err := <-errCh:
		return "", fmt.Errorf("serve oauth callback: %w", err)
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("timed out waiting for browser authentication")
		}
		return "", fmt.Errorf("wait for browser authentication: %w", ctx.Err())
	}
}

func (flow *loopbackAuthFlow) handleCallback(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", http.MethodGet)
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokenEndpoint, err := url.JoinPath(flow.baseURL, flow.tokenPath)
	if err != nil {
		http.Error(writer, "failed to build token endpoint", http.StatusInternalServerError)
		return
	}

	encodedTokenEndpoint, err := json.Marshal(tokenEndpoint)
	if err != nil {
		http.Error(writer, "failed to build token endpoint", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(
		writer,
		`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Finish sign-in</title>
  </head>
  <body data-token-endpoint="%s">
    <p id="status">Finishing sign-in...</p>
    <script>
      const tokenEndpoint = %s;
      const status = document.getElementById("status");
      const fragment = new URLSearchParams(window.location.hash.replace(/^#/, ""));
      const authToken = (fragment.get("auth_token") || "").trim();

      if (!authToken) {
        status.textContent = "Missing auth token. You can close this tab and try again.";
      } else {
        fetch(tokenEndpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ auth_token: authToken }),
        })
          .then((response) => {
            if (!response.ok) {
              throw new Error("token handoff failed");
            }
            status.textContent = "Authentication complete. You can close this tab.";
            window.setTimeout(() => window.close(), 100,);
          })
          .catch(() => {
            status.textContent = "Authentication handoff failed. Return to the terminal and try again.";
          });
      }
    </script>
  </body>
</html>`,
		html.EscapeString(tokenEndpoint),
		string(encodedTokenEndpoint),
	)
}

func (flow *loopbackAuthFlow) handleToken(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.Header().Set("Allow", http.MethodPost)
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer func() {
		_ = request.Body.Close()
	}()
	body, err := io.ReadAll(io.LimitReader(request.Body, 1<<20))
	if err != nil {
		http.Error(writer, "failed to read request body", http.StatusBadRequest)
		return
	}

	var payload struct {
		AuthToken string `json:"auth_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(payload.AuthToken)
	if token == "" {
		http.Error(writer, "missing auth token", http.StatusBadRequest)
		return
	}

	select {
	case flow.tokenCh <- token:
	default:
		http.Error(writer, "auth token already received", http.StatusConflict)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write([]byte(`{"status":"ok"}`))
}

func newLoopbackAuthNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
