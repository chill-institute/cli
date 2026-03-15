package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newCompletionCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:       "completion <shell>",
		Short:     "Generate shell completion scripts",
		Args:      allowDescribeArgs(cobra.ExactArgs(1)),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := strings.ToLower(strings.TrimSpace(args[0]))
			root := cmd.Root()

			switch shell {
			case "bash":
				return root.GenBashCompletion(app.stdout)
			case "zsh":
				return root.GenZshCompletion(app.stdout)
			case "fish":
				return root.GenFishCompletion(app.stdout, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(app.stdout)
			default:
				return usageError("unsupported_completion_shell", "unsupported shell %q (expected: bash, zsh, fish, powershell)", args[0])
			}
		},
		Example: strings.TrimSpace(`
chilly completion zsh > "${fpath[1]}/_chilly"
chilly completion bash > ~/.local/share/bash-completion/completions/chilly
chilly completion fish > ~/.config/fish/completions/chilly.fish
chilly completion powershell > chilly.ps1
`),
	}
}
