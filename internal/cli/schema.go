package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newSchemaCommand(app *appContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "schema",
		Short: "Inspect CLI command and procedure metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.writeJSONPayload(map[string]any{
				"commands":   listCommandSchemas(),
				"procedures": listProcedureSchemas(),
			})
		},
	}

	command.AddCommand(newSchemaCommandCommand(app))
	command.AddCommand(newSchemaProcedureCommand(app))
	return command
}

func newSchemaCommandCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "command <name>",
		Short: "Show metadata for one CLI command",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := lookupCommandSchema(strings.TrimSpace(args[0]))
			if !ok {
				return usageError("unknown_command_schema", "unknown command schema %q", args[0])
			}
			return app.writeJSONPayload(entry)
		},
	}
}

func newSchemaProcedureCommand(app *appContext) *cobra.Command {
	return &cobra.Command{
		Use:   "procedure <name>",
		Short: "Show metadata for one backend procedure",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := lookupProcedureSchema(strings.TrimSpace(args[0]))
			if !ok {
				return usageError("unknown_procedure_schema", "unknown procedure schema %q", args[0])
			}
			return app.writeJSONPayload(entry)
		},
	}
}

func allowDescribeArgs(next cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if describe, _ := cmd.Flags().GetBool("describe"); describe {
			return nil
		}
		return wrapUsageError("invalid_arguments", next(cmd, args))
	}
}
