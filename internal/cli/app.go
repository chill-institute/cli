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
	profile    string
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
	isTerminal      func(io.Writer) bool
	newTicker       func(time.Duration) progressTicker
	progressLabel   string
	progressEvery   time.Duration
}

type dryRunPreview struct {
	Status    string       `json:"status"`
	DryRun    bool         `json:"dry_run"`
	Command   string       `json:"command"`
	Procedure string       `json:"procedure,omitempty"`
	AuthMode  rpc.AuthMode `json:"auth_mode"`
	Request   any          `json:"request"`
}

type prettyRenderer func(any) (string, bool, error)

type progressTicker interface {
	C() <-chan time.Time
	Stop()
}

type realProgressTicker struct {
	ticker *time.Ticker
}

func (ticker realProgressTicker) C() <-chan time.Time {
	return ticker.ticker.C
}

func (ticker realProgressTicker) Stop() {
	ticker.ticker.Stop()
}

func newAppContext(opts *appOptions) *appContext {
	return &appContext{
		opts:            opts,
		stdin:           os.Stdin,
		stdout:          os.Stdout,
		stderr:          os.Stderr,
		openURL:         openBrowser,
		authFlowTimeout: 2 * time.Minute,
		isTerminal:      writerIsTerminal,
		newTicker: func(interval time.Duration) progressTicker {
			return realProgressTicker{ticker: time.NewTicker(interval)}
		},
		progressLabel: "Loading",
		progressEvery: 120 * time.Millisecond,
	}
}

func (app *appContext) configStore() (*config.Store, error) {
	configPath := strings.TrimSpace(app.opts.configPath)
	if configPath == "" {
		activeProfile, err := app.activeProfile()
		if err != nil {
			return nil, err
		}
		defaultPath, err := resolveDefaultConfigPath(activeProfile)
		if err != nil {
			return nil, wrapInternalError("config_path_resolve_failed", "resolve config path", err)
		}
		configPath = defaultPath
	}

	store, err := config.NewStore(configPath)
	if err != nil {
		return nil, wrapInternalError("config_store_init_failed", "initialize config store", err)
	}
	return store, nil
}

func (app *appContext) activeProfile() (string, error) {
	profile, err := resolveConfigProfile(app.opts.profile, currentBuildInfo().IsDev())
	if err != nil {
		return "", wrapUsageError("invalid_profile", err)
	}
	return profile, nil
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
	var response rpc.CallResponse
	err := app.withProgress(func() error {
		var callErr error
		response, callErr = app.rpcClient(cfg).Call(ctx, rpc.CallRequest{
			Procedure: procedure,
			Body:      body,
			AuthMode:  authMode,
			AuthToken: authToken,
		})
		return callErr
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

func (app *appContext) writeSelectedResponseBodyWithRenderer(body []byte, selection *fieldSelection, renderer prettyRenderer) error {
	if app.opts.output == outputJSON || selection != nil || renderer == nil {
		return app.writeSelectedResponseBody(body, selection)
	}

	var value any
	if err := json.Unmarshal(bytes.TrimSpace(body), &value); err != nil {
		return wrapInternalError("response_decode_failed", "decode response output", err)
	}

	rendered, ok, err := renderer(value)
	if err != nil {
		return wrapInternalError("response_render_failed", "render response output", err)
	}
	if !ok {
		return app.writeSelectedResponseBody(body, selection)
	}

	_, err = fmt.Fprintln(app.stdout, rendered)
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

func (app *appContext) writeAnyWithRenderer(payload any, selection *fieldSelection, renderer prettyRenderer) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return wrapInternalError("output_marshal_failed", "marshal output payload", err)
	}
	return app.writeSelectedResponseBodyWithRenderer(encoded, selection, renderer)
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

func (app *appContext) writeLocalDryRunPreview(commandID string, request any) error {
	return app.writeDryRunPreview(commandID, "", rpc.AuthNone, request)
}

func (app *appContext) readLine(prompt string) (string, error) {
	if _, err := fmt.Fprint(app.stderr, prompt); err != nil {
		return "", err
	}
	reader := bufio.NewReader(app.stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (app *appContext) withProgress(run func() error) error {
	if run == nil {
		return nil
	}
	if !app.shouldShowProgress() {
		return run()
	}

	ticker := app.newTicker(app.progressEvery)
	defer ticker.Stop()

	done := make(chan struct{})
	go app.renderProgress(done, ticker)

	err := run()
	close(done)
	app.clearProgress()
	return err
}

func (app *appContext) shouldShowProgress() bool {
	return app != nil &&
		app.opts != nil &&
		app.opts.output == outputPretty &&
		app.stderr != nil &&
		app.isTerminal != nil &&
		app.isTerminal(app.stderr)
}

func (app *appContext) renderProgress(done <-chan struct{}, ticker progressTicker) {
	frames := []string{"-", "\\", "|", "/"}
	frameIndex := 0

	for {
		if _, err := fmt.Fprintf(app.stderr, "\r%s %s", frames[frameIndex%len(frames)], app.progressLabel); err != nil {
			return
		}

		select {
		case <-done:
			return
		case <-ticker.C():
			frameIndex++
		}
	}
}

func (app *appContext) clearProgress() {
	if app == nil || app.stderr == nil {
		return
	}
	width := len(app.progressLabel) + 4
	_, _ = fmt.Fprintf(app.stderr, "\r%s\r", strings.Repeat(" ", width))
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

func writerIsTerminal(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
