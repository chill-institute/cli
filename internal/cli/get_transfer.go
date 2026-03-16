package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newGetTransferCommand(app *appContext) *cobra.Command {
	var fields string

	command := &cobra.Command{
		Use:   "get-transfer <id>",
		Short: "Show one transfer through chill.institute",
		Example: strings.TrimSpace(`
chilly get-transfer 42
chilly get-transfer 42 --output json
chilly get-transfer 42 --fields transfer.status,transfer.statusMessage,transfer.percentDone,transfer.fileId,transfer.fileUrl --output json
`),
		Args: allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := normalizeTransferID(args[0])
			if err != nil {
				return err
			}
			selection, err := parseFieldSelection(fields)
			if err != nil {
				return err
			}
			return runUserRPCWithRenderer(app, procedureUserGetTransfer, map[string]any{"id": id}, selection, renderTransferPretty)
		},
	}

	command.Flags().StringVar(&fields, "fields", "", "comma-separated field paths to include in the output")
	return command
}
