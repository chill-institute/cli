package cli

import (
	"io"
	"strings"

	"github.com/chill-institute/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	resolveConfigProfile     = config.ResolveProfile
	resolveDefaultConfigPath = config.DefaultPath
)

func NewRootCommand() *cobra.Command {
	opts := &appOptions{output: outputPretty}
	app := newAppContext(opts)
	return newRootCommand(app)
}

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runCommand(args, stdin, stdout, stderr)
}

func newRootCommand(app *appContext) *cobra.Command {
	opts := app.opts
	command := &cobra.Command{
		Use:   "chilly",
		Short: "chill.institute CLI for humans and agents",
		Example: strings.TrimSpace(`
chilly auth login
chilly search --query "dune"
chilly schema command search --output json
`),
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(opts.configPath) == "" {
				resolvedProfile, err := resolveConfigProfile(opts.profile, currentBuildInfo().IsDev())
				if err != nil {
					return usageError("invalid_profile", "invalid --profile %q", opts.profile)
				}
				opts.profile = resolvedProfile

				defaultPath, err := resolveDefaultConfigPath(resolvedProfile)
				if err != nil {
					return wrapInternalError("resolve_config_path_failed", "resolve config path", err)
				}
				opts.configPath = defaultPath
			} else {
				resolvedProfile, err := resolveConfigProfile(opts.profile, currentBuildInfo().IsDev())
				if err != nil {
					return usageError("invalid_profile", "invalid --profile %q", opts.profile)
				}
				opts.profile = resolvedProfile
			}
			opts.output = strings.ToLower(strings.TrimSpace(opts.output))
			if opts.output != outputPretty && opts.output != outputJSON {
				return usageError("invalid_output_mode", "invalid --output %q (expected: pretty|json)", opts.output)
			}

			describe, err := cmd.Flags().GetBool("describe")
			if err != nil {
				return wrapInternalError("describe_flag_lookup_failed", "resolve describe flag", err)
			}
			if describe {
				if err := describeCommand(app, cmd); err != nil {
					return err
				}
				cmd.RunE = func(*cobra.Command, []string) error { return nil }
				cmd.Run = nil
			}
			return nil
		},
	}

	command.SetIn(app.stdin)
	command.SetOut(app.stdout)
	command.SetErr(app.stderr)
	command.PersistentFlags().StringVar(&opts.configPath, "config", "", "config file path")
	command.PersistentFlags().StringVar(&opts.profile, "profile", "", "config profile to use (default builds use default, dev builds use dev)")
	command.PersistentFlags().StringVar(&opts.apiURL, "api-url", "", "override API base URL")
	command.PersistentFlags().StringVar(&opts.output, "output", outputPretty, "output mode: pretty|json")
	command.PersistentFlags().Bool("describe", false, "print command metadata and exit")

	command.AddCommand(newAuthCommand(app))
	command.AddCommand(newCompletionCommand(app))
	command.AddCommand(newDoctorCommand(app))
	command.AddCommand(newWhoamiCommand(app))
	command.AddCommand(newSettingsCommand(app))
	command.AddCommand(newSearchCommand(app))
	command.AddCommand(newListTopMoviesCommand(app))
	command.AddCommand(newAddTransferCommand(app))
	command.AddCommand(newUserCommand(app))
	command.AddCommand(newSchemaCommand(app))
	command.AddCommand(newVersionCommand(app))
	command.AddCommand(newSelfUpdateCommand(app))

	return command
}

func describeCommand(app *appContext, cmd *cobra.Command) error {
	entry, ok := lookupCommandSchema(schemaCommandID(cmd.CommandPath()))
	if !ok {
		return usageError("describe_not_supported", "describe is not supported for %q", cmd.CommandPath())
	}
	return app.writeJSONPayload(entry)
}
