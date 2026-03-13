package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
		Short: "Authenticate with a setup token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			resolvedToken := strings.TrimSpace(token)
			if resolvedToken == "" {
				loginURL := strings.TrimRight(cfg.APIBaseURL, "/") + "/auth/putio/start"
				if _, err := fmt.Fprintf(app.stdout, "Open this URL to authenticate:\n%s\n\n", loginURL); err != nil {
					return err
				}
				if !skipOpen {
					if err := app.openURL(loginURL); err != nil {
						_, _ = fmt.Fprintf(app.stderr, "Unable to open browser automatically: %v\n", err)
					}
				}

				inputToken, err := app.readLine("Paste setup token: ")
				if err != nil {
					return fmt.Errorf("read setup token: %w", err)
				}
				resolvedToken = strings.TrimSpace(inputToken)
			}

			if resolvedToken == "" {
				return fmt.Errorf("token cannot be empty")
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
	command.Flags().BoolVar(&skipOpen, "no-browser", false, "do not try to open browser for login flow")
	command.Flags().BoolVar(&skipVerify, "skip-verify", false, "skip token verification call")
	return command
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
