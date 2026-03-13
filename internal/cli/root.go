package cli

import (
	"fmt"
	"strings"

	"github.com/chill-institute/cli/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	opts := &appOptions{output: outputPretty}
	app := newAppContext(opts)

	command := &cobra.Command{
		Use:               "chilly",
		Short:             "Chill CLI for humans and agents",
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if strings.TrimSpace(opts.configPath) == "" {
				defaultPath, err := config.DefaultPath()
				if err != nil {
					return err
				}
				opts.configPath = defaultPath
			}
			opts.output = strings.ToLower(strings.TrimSpace(opts.output))
			if opts.output != outputPretty && opts.output != outputJSON {
				return fmt.Errorf("invalid --output %q (expected: pretty|json)", opts.output)
			}
			return nil
		},
	}

	command.PersistentFlags().StringVar(&opts.configPath, "config", "", "config file path")
	command.PersistentFlags().StringVar(&opts.apiURL, "api-url", "", "override API base URL")
	command.PersistentFlags().StringVar(&opts.output, "output", outputPretty, "output mode: pretty|json")

	command.AddCommand(newAuthCommand(app))
	command.AddCommand(newWhoamiCommand(app))
	command.AddCommand(newSettingsCommand(app))
	command.AddCommand(newSearchCommand(app))
	command.AddCommand(newListTopMoviesCommand(app))
	command.AddCommand(newAddTransferCommand(app))
	command.AddCommand(newUserCommand(app))

	return command
}
