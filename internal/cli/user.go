package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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
			return runUserRPCWithRenderer(app, procedureUserGetIndexers, map[string]any{}, nil, renderUserIndexersPretty)
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
	command.AddCommand(newUserDownloadFolderCommand(app))
	command.AddCommand(newUserFolderCommand(app))

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
		Use:   "set [field] [value]",
		Short: "Save full user settings JSON payload or patch one setting",
		Args:  allowDescribeArgs(cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmed := strings.TrimSpace(rawSettings)
			if trimmed != "" {
				if len(args) != 0 {
					return usageError("ambiguous_user_settings_update", "use either --json or <field> <value>, not both")
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
			}

			if len(args) != 2 {
				return usageError("missing_user_settings_update", "provide either --json or <field> <value>")
			}

			patch, err := normalizeUserSettingsPatch(args[0], args[1])
			if err != nil {
				return err
			}
			return runUserSettingsPatch(app, patch, dryRun)
		},
	}
	setCommand.Flags().StringVar(&rawSettings, "json", "", "full settings object JSON")
	setCommand.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the request or patch without executing it")
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

func newUserDownloadFolderCommand(app *appContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "download-folder",
		Short: "Show the current download folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserRPCWithRenderer(app, procedureUserGetDownloadFolder, map[string]any{}, nil, renderDownloadFolderPretty)
		},
	}

	var dryRun bool
	setCommand := &cobra.Command{
		Use:   "set <id>",
		Short: "Set the current download folder",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := normalizeFolderID(args[0])
			if err != nil {
				return err
			}
			return runUserSettingsPatch(app, userSettingsPatch{
				Field: "downloadFolderId",
				Value: strconv.FormatInt(id, 10),
			}, dryRun)
		},
	}
	setCommand.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the request or patch without executing it")
	command.AddCommand(setCommand)

	var clearDryRun bool
	clearCommand := &cobra.Command{
		Use:   "clear",
		Short: "Clear the current download folder setting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserSettingsPatch(app, userSettingsPatch{
				Field: "downloadFolderId",
				Value: nil,
			}, clearDryRun)
		},
	}
	clearCommand.Flags().BoolVar(&clearDryRun, "dry-run", false, "validate input and print the request or patch without executing it")
	command.AddCommand(clearCommand)

	return command
}

func newUserFolderCommand(app *appContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "folder",
		Short: "Folder operations",
	}

	getCommand := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one folder and its children",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := normalizeFolderID(args[0])
			if err != nil {
				return err
			}
			return runUserRPCWithRenderer(app, procedureUserGetFolder, map[string]any{"id": id}, nil, renderFolderPretty)
		},
	}

	command.AddCommand(getCommand)
	return command
}

func runUserRPC(app *appContext, procedure string, body any) error {
	return runUserRPCWithRenderer(app, procedure, body, nil, nil)
}

func loadCurrentUserSettings(app *appContext) (map[string]any, error) {
	cfg, err := app.loadConfig()
	if err != nil {
		return nil, err
	}
	token, err := app.userToken(cfg)
	if err != nil {
		return nil, err
	}

	response, err := app.callRPC(
		context.Background(),
		cfg,
		procedureUserGetUserSettings,
		map[string]any{},
		rpc.AuthUser,
		token,
	)
	if err != nil {
		return nil, fmt.Errorf("load current user settings: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(response.Body, &settings); err != nil {
		return nil, wrapInternalError("user_settings_decode_failed", "decode current user settings", err)
	}
	if settings == nil {
		settings = map[string]any{}
	}
	return settings, nil
}

func runUserSettingsPatch(app *appContext, patch userSettingsPatch, dryRun bool) error {
	if dryRun {
		return app.writeDryRunPreview("user settings set", procedureUserSaveUserSettings, rpc.AuthUser, map[string]any{
			"patch": patch,
		})
	}

	currentSettings, err := loadCurrentUserSettings(app)
	if err != nil {
		return err
	}
	request := map[string]any{"settings": applyUserSettingsPatch(currentSettings, patch)}
	return runUserRPC(app, procedureUserSaveUserSettings, request)
}

func normalizeFolderID(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, usageError("missing_folder_id", "folder id is required")
	}

	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, usageError("invalid_folder_id", "folder id must be an integer")
	}
	if value < 0 {
		return 0, usageError("invalid_folder_id", "folder id must be zero or positive")
	}
	return value, nil
}

func runUserRPCWithFields(app *appContext, procedure string, body any, selection *fieldSelection) error {
	return runUserRPCWithRenderer(app, procedure, body, selection, nil)
}

func runUserRPCWithRenderer(app *appContext, procedure string, body any, selection *fieldSelection, renderer prettyRenderer) error {
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
	if renderer == nil {
		renderer = prettyRendererForProcedure(procedure)
	}
	return app.writeSelectedResponseBodyWithRenderer(response.Body, selection, renderer)
}

func prettyRendererForProcedure(procedure string) prettyRenderer {
	switch procedure {
	case procedureUserGetIndexers:
		return renderUserIndexersPretty
	case procedureUserGetUserProfile:
		return renderWhoamiPretty
	case procedureUserGetUserSettings:
		return renderUserSettingsPretty
	case procedureUserSearch:
		return renderSearchPretty
	case procedureUserGetTopMovies:
		return renderTopMoviesPretty
	default:
		return nil
	}
}
