package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	AuthNone AuthMode = "none"
	AuthUser AuthMode = "user"
)

type AuthMode string

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type CallRequest struct {
	Procedure string
	Body      any
	AuthMode  AuthMode
	AuthToken string
}

type CallResponse struct {
	StatusCode int
	RequestID  string
	Body       json.RawMessage
}

type ErrorEnvelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Body       string
}

func (err APIError) Error() string {
	if err.Code != "" || err.Message != "" {
		return fmt.Sprintf("api error (%d): %s: %s", err.StatusCode, err.Code, err.Message)
	}
	return fmt.Sprintf("api error (%d): %s", err.StatusCode, err.Body)
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBaseURL == "" {
		trimmedBaseURL = "http://localhost:8080"
	}

	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	return &Client{
		baseURL:    trimmedBaseURL,
		httpClient: client,
	}
}

func (client Client) Call(ctx context.Context, req CallRequest) (CallResponse, error) {
	procedure, err := normalizeProcedure(req.Procedure)
	if err != nil {
		return CallResponse{}, err
	}

	endpoint, err := url.JoinPath(client.baseURL, "v4", procedure)
	if err != nil {
		return CallResponse{}, fmt.Errorf("build endpoint: %w", err)
	}

	payload, err := json.Marshal(defaultBody(req.Body))
	if err != nil {
		return CallResponse{}, fmt.Errorf("encode request body: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return CallResponse{}, fmt.Errorf("build request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("X-Request-Id", newRequestID())

	if err := applyAuth(httpRequest, req.AuthMode, req.AuthToken); err != nil {
		return CallResponse{}, err
	}

	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return CallResponse{}, fmt.Errorf("execute request: %w", err)
	}
	defer httpResponse.Body.Close()

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return CallResponse{}, fmt.Errorf("read response: %w", err)
	}

	callResponse := CallResponse{
		StatusCode: httpResponse.StatusCode,
		RequestID:  strings.TrimSpace(httpResponse.Header.Get("X-Request-Id")),
		Body:       responseBody,
	}

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return CallResponse{}, parseAPIError(callResponse)
	}

	return callResponse, nil
}

func normalizeProcedure(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("procedure is required")
	}
	trimmed = strings.TrimPrefix(trimmed, "/")
	if !strings.Contains(trimmed, "/") {
		return "", fmt.Errorf("invalid procedure %q", value)
	}
	return trimmed, nil
}

func defaultBody(body any) any {
	if body == nil {
		return map[string]any{}
	}
	return body
}

func applyAuth(request *http.Request, mode AuthMode, authToken string) error {
	switch mode {
	case AuthNone:
		return nil
	case AuthUser:
		trimmedToken := strings.TrimSpace(authToken)
		if trimmedToken == "" {
			return errors.New("missing auth token")
		}
		request.Header.Set("Authorization", "Bearer "+trimmedToken)
		return nil
	default:
		return fmt.Errorf("unsupported auth mode %q", mode)
	}
}

func parseAPIError(response CallResponse) error {
	apiErr := APIError{StatusCode: response.StatusCode, RequestID: response.RequestID, Body: strings.TrimSpace(string(response.Body))}

	var envelope ErrorEnvelope
	if err := json.Unmarshal(response.Body, &envelope); err == nil {
		apiErr.Code = strings.TrimSpace(envelope.Code)
		apiErr.Message = strings.TrimSpace(envelope.Message)
		if strings.TrimSpace(envelope.RequestID) != "" {
			apiErr.RequestID = strings.TrimSpace(envelope.RequestID)
		}
	}

	return apiErr
}

func newRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("cli-%d", time.Now().UnixNano())
	}
	return "cli-" + hex.EncodeToString(bytes)
}
