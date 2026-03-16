package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/chill-institute/chill-institute-cli/internal/rpc"
	"github.com/spf13/cobra"
)

func newAddTransferCommand(app *appContext) *cobra.Command {
	var transferURL string
	var dryRun bool

	command := &cobra.Command{
		Use:   "add-transfer",
		Short: "Add a transfer through chill.institute",
		Example: strings.TrimSpace(`
chilly add-transfer --url "magnet:?xt=urn:btih:..."
chilly add-transfer --url "magnet:?xt=urn:btih:..." --dry-run --output json
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAddTransfer(app, "add-transfer", transferURL, dryRun)
		},
	}

	command.Flags().StringVar(&transferURL, "url", "", "magnet or URL to add as transfer")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "validate input and print the request without executing it")
	return command
}

func runAddTransfer(app *appContext, commandID string, transferURL string, dryRun bool) error {
	trimmedURL := strings.TrimSpace(transferURL)
	if trimmedURL == "" {
		return usageError("missing_url", "--url is required")
	}
	request := map[string]any{"url": trimmedURL}
	if dryRun {
		return app.writeDryRunPreview(commandID, procedureUserAddTransfer, rpc.AuthUser, request)
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
	return app.writeSelectedResponseBodyWithRenderer(response.Body, nil, renderTransferPretty)
}
