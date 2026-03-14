package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chill-institute/cli/internal/config"
	"github.com/chill-institute/cli/internal/rpc"
	"github.com/spf13/cobra"
)

func newAuthCommand(app *appContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
	}

	command.AddCommand(newAuthLoginCommand(app))
	command.AddCommand(newAuthLogoutCommand(app))
	return command
}

func newAuthLoginCommand(app *appContext) *cobra.Command {
	var token string
	var skipOpen bool
	var skipVerify bool

	command := &cobra.Command{
		Use:   "login",
		Short: "Authenticate in a browser or store a setup token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			resolvedToken := strings.TrimSpace(token)
			if resolvedToken == "" {
				resolvedToken, err = app.loginWithBrowser(context.Background(), cfg, skipOpen)
				if err != nil {
					return err
				}
			}

			if resolvedToken == "" {
				return usageError("empty_token", "token cannot be empty")
			}

			var verifyResponse rpc.CallResponse
			if !skipVerify {
				verifyResponse, err = app.callRPC(
					context.Background(),
					cfg,
					procedureUserGetUserProfile,
					map[string]any{},
					rpc.AuthUser,
					resolvedToken,
				)
				if err != nil {
					return fmt.Errorf("verify auth token: %w", err)
				}
			}

			cfg.AuthToken = resolvedToken
			if err := app.saveConfig(cfg); err != nil {
				return fmt.Errorf("persist auth token: %w", err)
			}

			if skipVerify {
				return app.writeJSONPayload(map[string]any{"status": "ok", "saved": true})
			}
			return app.writeJSONPayload(map[string]any{
				"status":     "ok",
				"saved":      true,
				"request_id": verifyResponse.RequestID,
				"user":       jsonMessage(verifyResponse.Body),
			})
		},
	}

	command.Flags().StringVar(&token, "token", "", "setup token to store (non-interactive)")
	command.Flags().BoolVar(&skipOpen, "no-browser", false, "print the login URL instead of opening a browser automatically")
	command.Flags().BoolVar(&skipVerify, "skip-verify", false, "skip token verification call")
	return command
}

func (app *appContext) loginWithBrowser(ctx context.Context, cfg config.Config, skipOpen bool) (string, error) {
	flow, err := newLoopbackAuthFlow(cfg.APIBaseURL)
	if err != nil {
		return "", err
	}
	defer func() {
		if shutdownErr := flow.shutdown(); shutdownErr != nil {
			_, _ = fmt.Fprintf(app.stderr, "Unable to stop local auth server cleanly: %v\n", shutdownErr)
		}
	}()

	noticeWriter := app.stdout
	if app.opts.output == outputJSON {
		noticeWriter = app.stderr
	}
	if _, err := fmt.Fprintf(noticeWriter, "Open this URL to authenticate:\n%s\n\n", flow.loginURL); err != nil {
		return "", err
	}

	errCh := make(chan error, 1)
	flow.start(errCh)

	if !skipOpen {
		if err := app.openURL(flow.loginURL); err != nil {
			_, _ = fmt.Fprintf(app.stderr, "Unable to open browser automatically: %v\n", err)
		}
	}

	timeout := app.authFlowTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	token, err := flow.waitForToken(waitCtx, errCh)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(token), nil
}

func newAuthLogoutCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear auth token from local config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			cfg.AuthToken = ""
			if err := app.saveConfig(cfg); err != nil {
				return fmt.Errorf("persist config: %w", err)
			}
			return app.writeJSONPayload(map[string]any{"status": "ok", "logged_out": true})
		},
	}
}

func jsonMessage(raw []byte) any {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return map[string]any{"raw": string(raw)}
	}
	return value
}
