package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/chill-institute/cli/internal/rpc"
	"github.com/spf13/cobra"
)

func newSearchCommand(app *appContext) *cobra.Command {
	var query string
	var indexerID string

	command := &cobra.Command{
		Use:   "search",
		Short: "Search torrents using your user profile settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			trimmedQuery := strings.TrimSpace(query)
			if trimmedQuery == "" {
				return fmt.Errorf("--query is required")
			}

			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			token, err := app.userToken(cfg)
			if err != nil {
				return err
			}

			payload := map[string]any{"query": trimmedQuery}
			if trimmedIndexer := strings.TrimSpace(indexerID); trimmedIndexer != "" {
				payload["indexer_id"] = trimmedIndexer
			}

			response, err := app.callRPC(
				context.Background(),
				cfg,
				procedureUserSearch,
				payload,
				rpc.AuthUser,
				token,
			)
			if err != nil {
				return fmt.Errorf("search: %w", err)
			}
			return app.writeResponseBody(response.Body)
		},
	}

	command.Flags().StringVar(&query, "query", "", "search query")
	command.Flags().StringVar(&indexerID, "indexer-id", "", "optional indexer id")
	return command
}
