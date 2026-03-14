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

	var profileFields string
	profileCommand := &cobra.Command{
		Use:   "profile",
		Short: "Alias for whoami",
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseFieldSelection(profileFields)
			if err != nil {
				return err
			}
			return runUserRPCWithFields(app, procedureUserGetUserProfile, map[string]any{}, selection)
		},
	}
	profileCommand.Flags().StringVar(&profileFields, "fields", "", "comma-separated field paths to include in the output")
	command.AddCommand(profileCommand)

	command.AddCommand(&cobra.Command{
		Use:   "indexers",
		Short: "List user indexers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserRPC(app, procedureUserGetIndexers, map[string]any{})
		},
	})

	var query string
	var indexerID string
	var searchFields string
	searchCommand := &cobra.Command{
		Use:   "search",
		Short: "User search",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmedQuery := strings.TrimSpace(query)
			if trimmedQuery == "" {
				return usageError("missing_query", "--query is required")
			}
			selection, err := parseFieldSelection(searchFields)
			if err != nil {
				return err
			}
			payload := map[string]any{"query": trimmedQuery}
			if trimmedIndexer := strings.TrimSpace(indexerID); trimmedIndexer != "" {
				payload["indexer_id"] = trimmedIndexer
			}
			return runUserRPCWithFields(app, procedureUserSearch, payload, selection)
		},
	}
	searchCommand.Flags().StringVar(&query, "query", "", "search query")
	searchCommand.Flags().StringVar(&indexerID, "indexer-id", "", "optional indexer id")
	searchCommand.Flags().StringVar(&searchFields, "fields", "", "comma-separated field paths to include in the output")
	command.AddCommand(searchCommand)

	var topMoviesFields string
	topMoviesCommand := &cobra.Command{
		Use:   "top-movies",
		Short: "List top movies from user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseFieldSelection(topMoviesFields)
			if err != nil {
				return err
			}
			return runUserRPCWithFields(app, procedureUserGetTopMovies, map[string]any{}, selection)
		},
	}
	topMoviesCommand.Flags().StringVar(&topMoviesFields, "fields", "", "comma-separated field paths to include in the output")
	command.AddCommand(topMoviesCommand)

	command.AddCommand(newUserSettingsCommand(app))
	command.AddCommand(newUserTransferCommand(app))

	return command
}

func newUserSettingsCommand(app *appContext) *cobra.Command {
	settingsCommand := &cobra.Command{
		Use:   "settings",
		Short: "User settings operations",
	}

	var getFields string
	getCommand := &cobra.Command{
		Use:   "get",
		Short: "Fetch user settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseFieldSelection(getFields)
			if err != nil {
				return err
			}
			return runUserRPCWithFields(app, procedureUserGetUserSettings, map[string]any{}, selection)
		},
	}
	getCommand.Flags().StringVar(&getFields, "fields", "", "comma-separated field paths to include in the output")
	settingsCommand.AddCommand(getCommand)

	var rawSettings string
	var dryRun bool
	setCommand := &cobra.Command{
		Use:   "set",
		Short: "Save full user settings JSON payload",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmed := strings.TrimSpace(rawSettings)
			if trimmed == "" {
				return usageError("missing_json_payload", "--json is required")
			}

			var payload map[string]any
			if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
				return usageError("invalid_json_payload", "parse --json payload: %v", err)
			}
			request := map[string]any{"settings": payload}
			if dryRun {
				return app.writeDryRunPreview("user settings set", procedureUserSaveUserSettings, rpc.AuthUser, request)
			}
			return runUserRPC(app, procedureUserSaveUserSettings, request)
		},
	}
	setCommand.Flags().StringVar(&rawSettings, "json", "", "full settings object JSON")
	setCommand.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the request without executing it")
	settingsCommand.AddCommand(setCommand)

	return settingsCommand
}

func newUserTransferCommand(app *appContext) *cobra.Command {
	transferCommand := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer operations",
	}

	var transferURL string
	var dryRun bool
	addCommand := &cobra.Command{
		Use:   "add",
		Short: "Add transfer",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmed := strings.TrimSpace(transferURL)
			if trimmed == "" {
				return usageError("missing_url", "--url is required")
			}
			request := map[string]any{"url": trimmed}
			if dryRun {
				return app.writeDryRunPreview("user transfer add", procedureUserAddTransfer, rpc.AuthUser, request)
			}
			return runUserRPC(app, procedureUserAddTransfer, request)
		},
	}
	addCommand.Flags().StringVar(&transferURL, "url", "", "magnet or URL")
	addCommand.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the request without executing it")
	transferCommand.AddCommand(addCommand)

	return transferCommand
}

func runUserRPC(app *appContext, procedure string, body any) error {
	return runUserRPCWithFields(app, procedure, body, nil)
}

func runUserRPCWithFields(app *appContext, procedure string, body any, selection *fieldSelection) error {
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
	return app.writeSelectedResponseBody(response.Body, selection)
}
