package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/chill-institute/cli/internal/config"
	"github.com/spf13/cobra"
)

const redactedToken = "[redacted]"

func newSettingsCommand(app *appContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "settings",
		Short: "Manage local CLI config",
		Example: strings.TrimSpace(`
chilly settings show
chilly settings get api-base-url
chilly settings set api-base-url https://api.binge.institute --dry-run --output json
`),
	}

	command.AddCommand(newSettingsPathCommand(app))
	command.AddCommand(newSettingsShowCommand(app))
	command.AddCommand(newSettingsGetCommand(app))
	command.AddCommand(newSettingsSetCommand(app))
	return command
}

func newSettingsPathCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show local config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := app.configStore()
			if err != nil {
				return err
			}
			return app.writeJSONPayload(map[string]any{
				"path":    store.Path(),
				"profile": app.activeProfile(),
			})
		},
	}
}

func newSettingsShowCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show local CLI config (auth token redacted)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadStoredCLISettings(app)
			if err != nil {
				return err
			}
			authToken := ""
			if strings.TrimSpace(cfg.AuthToken) != "" {
				authToken = redactedToken
			}
			return app.writeJSONPayload(map[string]any{
				"profile":      app.activeProfile(),
				"api_base_url": cfg.APIBaseURL,
				"auth_token":   authToken,
			})
		},
	}
}

func newSettingsGetCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Show one local CLI setting",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := normalizeSettingsKey(args[0])
			if err != nil {
				return err
			}

			cfg, err := loadStoredCLISettings(app)
			if err != nil {
				return err
			}

			switch key {
			case "api-base-url":
				return app.writeJSONPayload(map[string]any{
					"key":   key,
					"value": cfg.APIBaseURL,
				})
			default:
				return fmt.Errorf("unsupported settings key %q", key)
			}
		},
	}
}

func newSettingsSetCommand(app *appContext) *cobra.Command {
	var dryRun bool

	command := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set one local CLI setting",
		Example: strings.TrimSpace(`
chilly settings set api-base-url https://api.binge.institute
chilly settings set api-base-url https://api.binge.institute --dry-run --output json
`),
		Args: allowDescribeArgs(cobra.ExactArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := normalizeSettingsKey(args[0])
			if err != nil {
				return err
			}

			cfg, err := loadStoredCLISettings(app)
			if err != nil {
				return err
			}

			request := map[string]any{
				"key": args[0],
			}

			switch key {
			case "api-base-url":
				nextValue, err := normalizeAPIBaseURL(args[1])
				if err != nil {
					return err
				}
				request["key"] = key
				request["value"] = nextValue
				if dryRun {
					return app.writeLocalDryRunPreview("settings set", request)
				}
				cfg.APIBaseURL = nextValue
			default:
				return fmt.Errorf("unsupported settings key %q", key)
			}

			if err := app.saveConfig(cfg); err != nil {
				return err
			}

			return app.writeJSONPayload(map[string]any{
				"status": "ok",
				"key":    key,
				"value":  cfg.APIBaseURL,
			})
		},
	}

	command.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the local config change without saving it")
	return command
}

func loadStoredCLISettings(app *appContext) (config.Config, error) {
	store, err := app.configStore()
	if err != nil {
		return config.Config{}, err
	}
	cfg, err := store.Load()
	if err != nil {
		return config.Config{}, err
	}
	return cfg.Normalized(), nil
}

func normalizeSettingsKey(raw string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "api-base-url", "api_base_url":
		return "api-base-url", nil
	default:
		return "", usageError("unsupported_settings_key", "unsupported settings key %q (supported: api-base-url)", raw)
	}
}

func normalizeAPIBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", usageError("empty_api_base_url", "api-base-url cannot be empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", usageError("invalid_api_base_url", "parse api-base-url: %v", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", usageError("invalid_api_base_url", "api-base-url must include scheme and host")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", usageError("invalid_api_base_url", "api-base-url must start with http:// or https://")
	}

	return strings.TrimRight(trimmed, "/"), nil
}
