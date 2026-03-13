package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chill-institute/cli/internal/rpc"
	"github.com/spf13/cobra"
)

func newUserCommand(app *appContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "user",
		Short: "User RPC commands (Bearer auth)",
	}

	command.AddCommand(&cobra.Command{
		Use:   "profile",
		Short: "Alias for whoami",
		RunE: func(cmd *cobra.Command, args []string) error {
			return newWhoamiCommand(app).RunE(cmd, args)
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "indexers",
		Short: "List user indexers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserRPC(app, procedureUserGetIndexers, map[string]any{})
		},
	})

	var query string
	var indexerID string
	searchCommand := &cobra.Command{
		Use:   "search",
		Short: "User search",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmedQuery := strings.TrimSpace(query)
			if trimmedQuery == "" {
				return fmt.Errorf("--query is required")
			}
			payload := map[string]any{"query": trimmedQuery}
			if trimmedIndexer := strings.TrimSpace(indexerID); trimmedIndexer != "" {
				payload["indexer_id"] = trimmedIndexer
			}
			return runUserRPC(app, procedureUserSearch, payload)
		},
	}
	searchCommand.Flags().StringVar(&query, "query", "", "search query")
	searchCommand.Flags().StringVar(&indexerID, "indexer-id", "", "optional indexer id")
	command.AddCommand(searchCommand)

	command.AddCommand(&cobra.Command{
		Use:   "top-movies",
		Short: "List top movies from user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserRPC(app, procedureUserGetTopMovies, map[string]any{})
		},
	})

	command.AddCommand(newUserSettingsCommand(app))
	command.AddCommand(newUserTransferCommand(app))

	return command
}

func newUserSettingsCommand(app *appContext) *cobra.Command {
	settingsCommand := &cobra.Command{
		Use:   "settings",
		Short: "User settings operations",
	}

	settingsCommand.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Fetch user settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserRPC(app, procedureUserGetUserSettings, map[string]any{})
		},
	})

	var rawSettings string
	setCommand := &cobra.Command{
		Use:   "set",
		Short: "Save full user settings JSON payload",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmed := strings.TrimSpace(rawSettings)
			if trimmed == "" {
				return fmt.Errorf("--json is required")
			}

			var payload map[string]any
			if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
				return fmt.Errorf("parse --json payload: %w", err)
			}
			return runUserRPC(app, procedureUserSaveUserSettings, map[string]any{"settings": payload})
		},
	}
	setCommand.Flags().StringVar(&rawSettings, "json", "", "full settings object JSON")
	settingsCommand.AddCommand(setCommand)

	return settingsCommand
}

func newUserTransferCommand(app *appContext) *cobra.Command {
	transferCommand := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer operations",
	}

	var transferURL string
	addCommand := &cobra.Command{
		Use:   "add",
		Short: "Add transfer",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmed := strings.TrimSpace(transferURL)
			if trimmed == "" {
				return fmt.Errorf("--url is required")
			}
			return runUserRPC(app, procedureUserAddTransfer, map[string]any{"url": trimmed})
		},
	}
	addCommand.Flags().StringVar(&transferURL, "url", "", "magnet or URL")
	transferCommand.AddCommand(addCommand)

	return transferCommand
}

func runUserRPC(app *appContext, procedure string, body any) error {
	cfg, err := app.loadConfig()
	if err != nil {
		return err
	}
	token, err := app.userToken(cfg)
	if err != nil {
		return err
	}

	response, err := app.callRPC(
		context.Background(),
		cfg,
		procedure,
		body,
		rpc.AuthUser,
		token,
	)
	if err != nil {
		return fmt.Errorf("user rpc call: %w", err)
	}
	return app.writeResponseBody(response.Body)
}
