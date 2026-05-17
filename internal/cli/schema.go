package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func newSchemaCommand(app *appContext) *cobra.Command {
	var fields string
	command := &cobra.Command{
		Use:   "schema",
		Short: "Inspect command and procedure contracts",
		Example: strings.TrimSpace(`
chilly schema --output json
chilly schema command search --output json
chilly search --describe --output json
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			selection, err := parseFieldSelection(fields)
			if err != nil {
				return err
			}
			return app.writeAnyWithRenderer(map[string]any{
				"commands":   listCommandSchemas(),
				"procedures": listProcedureSchemas(),
				"types":      listTypeSchemas(),
			}, selection, nil)
		},
	}

	command.Flags().StringVar(&fields, "fields", "", "comma-separated field paths to include in the output")
	command.AddCommand(newSchemaCommandCommand(app))
	command.AddCommand(newSchemaProcedureCommand(app))
	command.AddCommand(newSchemaTypeCommand(app))
	return command
}

func newSchemaCommandCommand(app *appContext) *cobra.Command {
	var fields string
	command := &cobra.Command{
		Use:   "command <name>",
		Short: "Show metadata for one CLI command",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := lookupCommandSchema(strings.TrimSpace(args[0]))
			if !ok {
				return usageError("unknown_command_schema", "unknown command schema %q", args[0])
			}
			selection, err := parseFieldSelection(fields)
			if err != nil {
				return err
			}
			return app.writeAnyWithRenderer(entry, selection, nil)
		},
	}
	command.Flags().StringVar(&fields, "fields", "", "comma-separated field paths to include in the output")
	return command
}

func newSchemaProcedureCommand(app *appContext) *cobra.Command {
	var fields string
	command := &cobra.Command{
		Use:   "procedure <name>",
		Short: "Show metadata for one backend procedure",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := lookupProcedureSchema(strings.TrimSpace(args[0]))
			if !ok {
				return usageError("unknown_procedure_schema", "unknown procedure schema %q", args[0])
			}
			selection, err := parseFieldSelection(fields)
			if err != nil {
				return err
			}
			return app.writeAnyWithRenderer(entry, selection, nil)
		},
	}
	command.Flags().StringVar(&fields, "fields", "", "comma-separated field paths to include in the output")
	return command
}

func newSchemaTypeCommand(app *appContext) *cobra.Command {
	var fields string
	command := &cobra.Command{
		Use:   "type <name>",
		Short: "Show metadata for one output type",
		Args:  allowDescribeArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := lookupTypeSchema(strings.TrimSpace(args[0]))
			if !ok {
				return usageError("unknown_type_schema", "unknown type schema %q", args[0])
			}
			selection, err := parseFieldSelection(fields)
			if err != nil {
				return err
			}
			return app.writeAnyWithRenderer(entry, selection, nil)
		},
	}
	command.Flags().StringVar(&fields, "fields", "", "comma-separated field paths to include in the output")
	return command
}

func allowDescribeArgs(next cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if describe, _ := cmd.Flags().GetBool("describe"); describe {
			return nil
		}
		return wrapUsageError("invalid_arguments", next(cmd, args))
	}
}
