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
		Short: "Run user account commands through chill.institute",
		Example: strings.TrimSpace(`
chilly user profile
chilly user settings get --output json
chilly user download-folder
`),
	}

	var profileFields string
	profileCommand := &cobra.Command{
		Use:   "profile",
		Short: "Show authenticated profile (alias for whoami)",
		Long:  "Alias for the top-level whoami command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami(app, profileFields)
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
		Short: "Search using your saved profile settings",
		Long:  "Alias for the top-level search command under the user namespace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(app, query, indexerID, searchFields)
		},
	}
	searchCommand.Flags().StringVar(&query, "query", "", "search query")
	searchCommand.Flags().StringVar(&indexerID, "indexer-id", "", "optional indexer id")
	searchCommand.Flags().StringVar(&searchFields, "fields", "", "comma-separated field paths to include in the output")
	command.AddCommand(searchCommand)

	var topMoviesFields string
	topMoviesCommand := &cobra.Command{
		Use:   "top-movies",
		Short: "List top movies using your profile settings",
		Long:  "Alias for the top-level list-top-movies command.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListTopMovies(app, topMoviesFields)
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
		Short: "Read and update hosted user settings",
	}

	var getFields string
	getCommand := &cobra.Command{
		Use:   "get",
		Short: "Show hosted user settings",
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
		Long: strings.TrimSpace(`
Save user settings in one of two modes:

- full replacement with --json
- single-field patching with <field> <value>

` + supportedUserSettingsPatchHelp()),
		Example: strings.TrimSpace(`
chilly user settings set show-top-movies true
chilly user settings set sort-by title --dry-run --output json
chilly user settings set --json '{"showTopMovies":true}'
`),
		Args: allowDescribeArgs(cobra.MaximumNArgs(2)),
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
			return runUserSettingsPatch(app, "user settings set", patch, dryRun)
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
		Short: "Transfer commands",
	}

	var transferURL string
	var dryRun bool
	addCommand := &cobra.Command{
		Use:   "add",
		Short: "Add a transfer through chill.institute",
		Long:  "Alias for the top-level add-transfer command under the user namespace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddTransfer(app, "user transfer add", transferURL, dryRun)
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
		Short: "Show your current download folder",
		Example: strings.TrimSpace(`
chilly user download-folder
chilly user download-folder set 42 --dry-run --output json
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserRPCWithRenderer(app, procedureUserGetDownloadFolder, map[string]any{}, nil, renderDownloadFolderPretty)
		},
	}

	var dryRun bool
	setCommand := &cobra.Command{
		Use:   "set <id>",
		Short: "Set the current download folder",
		Example: strings.TrimSpace(`
chilly user download-folder set 42
chilly user download-folder set 42 --dry-run --output json
`),
		Args: allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := normalizeFolderID(args[0])
			if err != nil {
				return err
			}
			return runUserSettingsPatch(app, "user download-folder set", userSettingsPatch{
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
		Example: strings.TrimSpace(`
chilly user download-folder clear
chilly user download-folder clear --dry-run --output json
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserSettingsPatch(app, "user download-folder clear", userSettingsPatch{
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
		Short: "Inspect folders",
	}

	getCommand := &cobra.Command{
		Use:   "get <id>",
		Short: "Get one folder and its children",
		Example: strings.TrimSpace(`
chilly user folder get 0
chilly user folder get 42 --output json
`),
		Args: allowDescribeArgs(cobra.ExactArgs(1)),
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

func runUserSettingsPatch(app *appContext, commandID string, patch userSettingsPatch, dryRun bool) error {
	if dryRun {
		return app.writeDryRunPreview(commandID, procedureUserSaveUserSettings, rpc.AuthUser, map[string]any{
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
