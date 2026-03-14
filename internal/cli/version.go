package cli

import (
	"github.com/spf13/cobra"
)

func newVersionCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show CLI build metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := currentBuildInfo()
			return app.writeJSONPayload(map[string]any{
				"name":       "chilly",
				"version":    info.Version,
				"commit":     info.Commit,
				"build_date": info.BuildDate,
			})
		},
	}
}
