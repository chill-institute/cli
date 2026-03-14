package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/chill-institute/cli/internal/config"
	"github.com/chill-institute/cli/internal/rpc"
)

const (
	outputPretty = "pretty"
	outputJSON   = "json"
)

type appOptions struct {
	configPath string
	apiURL     string
	output     string
}

type appContext struct {
	opts            *appOptions
	stdin           io.Reader
	stdout          io.Writer
	stderr          io.Writer
	openURL         func(string) error
	authFlowTimeout time.Duration
}

type dryRunPreview struct {
	Status    string       `json:"status"`
	DryRun    bool         `json:"dry_run"`
	Command   string       `json:"command"`
	Procedure string       `json:"procedure"`
	AuthMode  rpc.AuthMode `json:"auth_mode"`
	Request   any          `json:"request"`
}

func newAppContext(opts *appOptions) *appContext {
	return &appContext{
		opts:            opts,
		stdin:           os.Stdin,
		stdout:          os.Stdout,
		stderr:          os.Stderr,
		openURL:         openBrowser,
		authFlowTimeout: 2 * time.Minute,
	}
}

func (app *appContext) configStore() (*config.Store, error) {
	store, err := config.NewStore(app.opts.configPath)
	if err != nil {
		return nil, wrapInternalError("config_store_init_failed", "initialize config store", err)
	}
	return store, nil
}

func (app *appContext) loadConfig() (config.Config, error) {
	store, err := app.configStore()
	if err != nil {
		return config.Config{}, err
	}
	cfg, err := store.Load()
	if err != nil {
		return config.Config{}, wrapInternalError("config_load_failed", "load config", err)
	}
	if override := strings.TrimSpace(app.opts.apiURL); override != "" {
		cfg.APIBaseURL = override
	}
	return cfg.Normalized(), nil
}

func (app *appContext) saveConfig(cfg config.Config) error {
	store, err := app.configStore()
	if err != nil {
		return err
	}
	return wrapInternalError("config_save_failed", "save config", store.Save(cfg))
}

func (app *appContext) rpcClient(cfg config.Config) *rpc.Client {
	return rpc.NewClient(cfg.APIBaseURL, http.DefaultClient)
}

func (app *appContext) userToken(cfg config.Config) (string, error) {
	token := strings.TrimSpace(cfg.AuthToken)
	if token == "" {
		return "", authError("missing_auth_token", "missing auth token: run `chilly auth login`")
	}
	return token, nil
}

func (app *appContext) callRPC(
	ctx context.Context,
	cfg config.Config,
	procedure string,
	body any,
	authMode rpc.AuthMode,
	authToken string,
) (rpc.CallResponse, error) {
	response, err := app.rpcClient(cfg).Call(ctx, rpc.CallRequest{
		Procedure: procedure,
		Body:      body,
		AuthMode:  authMode,
		AuthToken: authToken,
	})
	if err != nil {
		return rpc.CallResponse{}, err
	}
	return response, nil
}

func (app *appContext) writeResponseBody(body []byte) error {
	normalized, err := normalizeJSON(body, app.opts.output, nil)
	if err != nil {
		return wrapInternalError("response_normalize_failed", "normalize response output", err)
	}
	_, err = fmt.Fprintln(app.stdout, string(normalized))
	return wrapInternalError("stdout_write_failed", "write response output", err)
}

func (app *appContext) writeSelectedResponseBody(body []byte, selection *fieldSelection) error {
	normalized, err := normalizeJSON(body, app.opts.output, selection)
	if err != nil {
		return wrapInternalError("response_normalize_failed", "normalize response output", err)
	}
	_, err = fmt.Fprintln(app.stdout, string(normalized))
	return wrapInternalError("stdout_write_failed", "write response output", err)
}

func (app *appContext) writeJSONPayload(payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return wrapInternalError("output_marshal_failed", "marshal output payload", err)
	}
	normalized, err := normalizeJSON(encoded, app.opts.output, nil)
	if err != nil {
		return wrapInternalError("output_normalize_failed", "normalize output payload", err)
	}
	_, err = fmt.Fprintln(app.stdout, string(normalized))
	return wrapInternalError("stdout_write_failed", "write output payload", err)
}

func (app *appContext) writeDryRunPreview(commandID string, procedure string, authMode rpc.AuthMode, request any) error {
	return app.writeJSONPayload(dryRunPreview{
		Status:    "ok",
		DryRun:    true,
		Command:   commandID,
		Procedure: procedure,
		AuthMode:  authMode,
		Request:   request,
	})
}

func (app *appContext) readLine(prompt string) (string, error) {
	if _, err := fmt.Fprint(app.stdout, prompt); err != nil {
		return "", err
	}
	reader := bufio.NewReader(app.stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func normalizeJSON(raw []byte, mode string, selection *fieldSelection) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return []byte("{}"), nil
	}

	var value any
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, fmt.Errorf("invalid json response: %w", err)
	}
	value = selection.apply(value)

	if mode == outputJSON {
		return json.Marshal(value)
	}
	return json.MarshalIndent(value, "", "  ")
}

func openBrowser(rawURL string) error {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return fmt.Errorf("empty URL")
	}

	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", trimmedURL)
	case "windows":
		command = exec.Command("cmd", "/c", "start", "", trimmedURL)
	default:
		command = exec.Command("xdg-open", trimmedURL)
	}

	if err := command.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
