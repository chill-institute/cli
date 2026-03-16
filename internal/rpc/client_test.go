package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallUserAuthAndRequestBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %q, want %q", request.Method, http.MethodPost)
		}
		if request.URL.Path != "/v4/chill.v4.UserService/GetUserProfile" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q", got)
		}
		if request.Header.Get("X-Request-Id") == "" {
			t.Fatal("expected X-Request-Id header")
		}

		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if len(payload) != 1 || payload["query"] != "movie" {
			t.Fatalf("payload = %#v", payload)
		}

		writer.Header().Set("X-Request-Id", "req-123")
		_, _ = writer.Write([]byte(`{"user_id":"user-1"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	response, err := client.Call(context.Background(), CallRequest{
		Procedure: "/chill.v4.UserService/GetUserProfile",
		Body:      map[string]any{"query": "movie"},
		AuthMode:  AuthUser,
		AuthToken: "test-token",
	})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
	if response.RequestID != "req-123" {
		t.Fatalf("RequestID = %q, want %q", response.RequestID, "req-123")
	}
	if strings.TrimSpace(string(response.Body)) != `{"user_id":"user-1"}` {
		t.Fatalf("Body = %s", string(response.Body))
	}
}

func TestCallNoAuthHeaderWhenModeNone(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty", got)
		}
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	_, err := client.Call(context.Background(), CallRequest{
		Procedure: "chill.v4.UserService/GetUserProfile",
		AuthMode:  AuthNone,
	})
	if err != nil {
		t.Fatalf("Call() error = %v", err)
	}
}

func TestCallReturnsStructuredAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write([]byte(`{"code":"invalid_auth_token","message":"invalid auth token","request_id":"req-500"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, server.Client())
	_, err := client.Call(context.Background(), CallRequest{
		Procedure: "chill.v4.UserService/GetUserProfile",
		AuthMode:  AuthUser,
		AuthToken: "bad-token",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Code != "invalid_auth_token" {
		t.Fatalf("code = %q", apiErr.Code)
	}
	if apiErr.RequestID != "req-500" {
		t.Fatalf("request id = %q", apiErr.RequestID)
	}
}

func TestCallMissingUserToken(t *testing.T) {
	t.Parallel()

	client := NewClient("https://api.chill.institute", http.DefaultClient)
	_, err := client.Call(context.Background(), CallRequest{
		Procedure: "chill.v4.UserService/GetUserProfile",
		AuthMode:  AuthUser,
	})
	if err == nil {
		t.Fatal("expected missing token error")
	}
	if !strings.Contains(err.Error(), "missing auth token") {
		t.Fatalf("error = %v", err)
	}
}

func TestAPIErrorErrorIncludesBodyWhenEnvelopeMissing(t *testing.T) {
	t.Parallel()

	err := APIError{
		StatusCode: http.StatusBadGateway,
		Body:       "upstream failed",
	}

	if got := err.Error(); got != "api error (502): upstream failed" {
		t.Fatalf("Error() = %q", got)
	}
}
