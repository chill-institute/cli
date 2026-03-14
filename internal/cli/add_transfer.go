package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/chill-institute/cli/internal/rpc"
	"github.com/spf13/cobra"
)

func newAddTransferCommand(app *appContext) *cobra.Command {
	var transferURL string
	var dryRun bool

	command := &cobra.Command{
		Use:   "add-transfer",
		Short: "Add transfer to put.io",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmedURL := strings.TrimSpace(transferURL)
			if trimmedURL == "" {
				return usageError("missing_url", "--url is required")
			}
			request := map[string]any{"url": trimmedURL}
			if dryRun {
				return app.writeDryRunPreview("add-transfer", procedureUserAddTransfer, rpc.AuthUser, request)
			}

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
				procedureUserAddTransfer,
				request,
				rpc.AuthUser,
				token,
			)
			if err != nil {
				return fmt.Errorf("add transfer: %w", err)
			}
			return app.writeResponseBody(response.Body)
		},
	}

	command.Flags().StringVar(&transferURL, "url", "", "magnet or URL to add as transfer")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the request without executing it")
	return command
}
