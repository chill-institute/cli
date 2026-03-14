package cli

import (
	"context"
	"fmt"

	"github.com/chill-institute/cli/internal/rpc"
	"github.com/spf13/cobra"
)

func newWhoamiCommand(app *appContext) *cobra.Command {
	var fields string

	command := &cobra.Command{
		Use:   "whoami",
		Short: "Show authenticated user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseFieldSelection(fields)
			if err != nil {
				return err
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
				procedureUserGetUserProfile,
				map[string]any{},
				rpc.AuthUser,
				token,
			)
			if err != nil {
				return fmt.Errorf("get user profile: %w", err)
			}
			return app.writeSelectedResponseBody(response.Body, selection)
		},
	}

	command.Flags().StringVar(&fields, "fields", "", "comma-separated field paths to include in the output")
	return command
}
