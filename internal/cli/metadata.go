package cli

import (
	"sort"
	"strings"
)

type schemaInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type schemaOutput struct {
	JSON  bool `json:"json"`
	Human bool `json:"human"`
}

type schemaEntry struct {
	ID              string        `json:"id"`
	Kind            string        `json:"kind"`
	Summary         string        `json:"summary"`
	AuthMode        string        `json:"auth_mode"`
	Mutates         bool          `json:"mutates"`
	LinkedProcedure string        `json:"linked_procedure,omitempty"`
	Inputs          []schemaInput `json:"inputs,omitempty"`
	Output          schemaOutput  `json:"output"`
}

var commonCommandInputs = []schemaInput{
	{Name: "api-url", Type: "string", Description: "override API base URL"},
	{Name: "config", Type: "string", Description: "config file path"},
	{Name: "output", Type: "string", Description: "output mode: pretty|json"},
	{Name: "describe", Type: "boolean", Description: "print command metadata and exit"},
}

var commandSchemaRegistry = map[string]schemaEntry{
	"chilly": {
		ID:       "chilly",
		Kind:     "command",
		Summary:  "Chill CLI for humans and agents",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"add-transfer": {
		ID:              "add-transfer",
		Kind:            "command",
		Summary:         "Add transfer to put.io",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		LinkedProcedure: procedureUserAddTransfer,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "url", Type: "string", Required: true, Description: "magnet or URL to add as transfer"},
		),
	},
	"auth": {
		ID:       "auth",
		Kind:     "command",
		Summary:  "Authentication commands",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"auth login": {
		ID:       "auth login",
		Kind:     "command",
		Summary:  "Authenticate in a browser or store a setup token",
		AuthMode: string(rpcAuthNone),
		Mutates:  true,
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "token", Type: "string", Description: "setup token to store directly"},
			schemaInput{Name: "no-browser", Type: "boolean", Description: "print the login URL instead of opening a browser automatically"},
			schemaInput{Name: "skip-verify", Type: "boolean", Description: "skip token verification call"},
		),
	},
	"auth logout": {
		ID:       "auth logout",
		Kind:     "command",
		Summary:  "Clear auth token from local config",
		AuthMode: string(rpcAuthNone),
		Mutates:  true,
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"list-top-movies": {
		ID:              "list-top-movies",
		Kind:            "command",
		Summary:         "List top movies for your profile",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserGetTopMovies,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs:          cloneInputs(commonCommandInputs),
	},
	"schema": {
		ID:       "schema",
		Kind:     "command",
		Summary:  "Inspect CLI command and procedure metadata",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"schema command": {
		ID:       "schema command",
		Kind:     "command",
		Summary:  "Show metadata for one CLI command",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "name", Type: "string", Required: true, Description: "command id such as search or settings get"},
		),
	},
	"schema procedure": {
		ID:       "schema procedure",
		Kind:     "command",
		Summary:  "Show metadata for one backend procedure",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "procedure", Type: "string", Required: true, Description: "procedure id such as chill.v4.UserService/Search"},
		),
	},
	"search": {
		ID:              "search",
		Kind:            "command",
		Summary:         "Search torrents using your user profile settings",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserSearch,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "query", Type: "string", Required: true, Description: "search query"},
			schemaInput{Name: "indexer-id", Type: "string", Description: "optional indexer id"},
		),
	},
	"self-update": {
		ID:       "self-update",
		Kind:     "command",
		Summary:  "Download and install the latest released CLI binary",
		AuthMode: string(rpcAuthNone),
		Mutates:  true,
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "check", Type: "boolean", Description: "check for a newer release without installing it"},
			schemaInput{Name: "version", Type: "string", Description: "specific release tag to install"},
		),
	},
	"settings": {
		ID:       "settings",
		Kind:     "command",
		Summary:  "Manage local CLI settings",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"settings get": {
		ID:       "settings get",
		Kind:     "command",
		Summary:  "Get a local CLI setting value",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "key", Type: "string", Required: true, Description: "settings key such as api-base-url"},
		),
	},
	"settings path": {
		ID:       "settings path",
		Kind:     "command",
		Summary:  "Print local settings file path",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"settings set": {
		ID:       "settings set",
		Kind:     "command",
		Summary:  "Set a local CLI setting value",
		AuthMode: string(rpcAuthNone),
		Mutates:  true,
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "key", Type: "string", Required: true, Description: "settings key such as api-base-url"},
			schemaInput{Name: "value", Type: "string", Required: true, Description: "next value for the selected settings key"},
		),
	},
	"settings show": {
		ID:       "settings show",
		Kind:     "command",
		Summary:  "Show local CLI settings (auth token redacted)",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user": {
		ID:       "user",
		Kind:     "command",
		Summary:  "User RPC commands (Bearer auth)",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user indexers": {
		ID:              "user indexers",
		Kind:            "command",
		Summary:         "List user indexers",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserGetIndexers,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs:          cloneInputs(commonCommandInputs),
	},
	"user profile": {
		ID:              "user profile",
		Kind:            "command",
		Summary:         "Alias for whoami",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserGetUserProfile,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs:          cloneInputs(commonCommandInputs),
	},
	"user search": {
		ID:              "user search",
		Kind:            "command",
		Summary:         "User search",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserSearch,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "query", Type: "string", Required: true, Description: "search query"},
			schemaInput{Name: "indexer-id", Type: "string", Description: "optional indexer id"},
		),
	},
	"user settings": {
		ID:       "user settings",
		Kind:     "command",
		Summary:  "User settings operations",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user settings get": {
		ID:              "user settings get",
		Kind:            "command",
		Summary:         "Fetch user settings",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserGetUserSettings,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs:          cloneInputs(commonCommandInputs),
	},
	"user settings set": {
		ID:              "user settings set",
		Kind:            "command",
		Summary:         "Save full user settings JSON payload",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		LinkedProcedure: procedureUserSaveUserSettings,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "json", Type: "string", Required: true, Description: "full settings object JSON"},
		),
	},
	"user top-movies": {
		ID:              "user top-movies",
		Kind:            "command",
		Summary:         "List top movies from user profile",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserGetTopMovies,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs:          cloneInputs(commonCommandInputs),
	},
	"user transfer": {
		ID:       "user transfer",
		Kind:     "command",
		Summary:  "Transfer operations",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user transfer add": {
		ID:              "user transfer add",
		Kind:            "command",
		Summary:         "Add transfer",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		LinkedProcedure: procedureUserAddTransfer,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "url", Type: "string", Required: true, Description: "magnet or URL"},
		),
	},
	"whoami": {
		ID:              "whoami",
		Kind:            "command",
		Summary:         "Show authenticated user profile",
		AuthMode:        string(rpcAuthUser),
		LinkedProcedure: procedureUserGetUserProfile,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs:          cloneInputs(commonCommandInputs),
	},
	"version": {
		ID:       "version",
		Kind:     "command",
		Summary:  "Show CLI build metadata",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
}

var procedureSchemaRegistry = map[string]schemaEntry{
	procedureUserAddTransfer: {
		ID:       procedureUserAddTransfer,
		Kind:     "procedure",
		Summary:  "Add transfer to put.io",
		AuthMode: string(rpcAuthUser),
		Mutates:  true,
		Output:   schemaOutput{JSON: true},
		Inputs: []schemaInput{
			{Name: "url", Type: "string", Required: true, Description: "magnet or URL to add as transfer"},
		},
	},
	procedureUserGetIndexers: {
		ID:       procedureUserGetIndexers,
		Kind:     "procedure",
		Summary:  "List user indexers",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true},
	},
	procedureUserGetTopMovies: {
		ID:       procedureUserGetTopMovies,
		Kind:     "procedure",
		Summary:  "List top movies for the current user",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true},
	},
	procedureUserGetUserProfile: {
		ID:       procedureUserGetUserProfile,
		Kind:     "procedure",
		Summary:  "Fetch the authenticated user profile",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true},
	},
	procedureUserGetUserSettings: {
		ID:       procedureUserGetUserSettings,
		Kind:     "procedure",
		Summary:  "Fetch user settings",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true},
	},
	procedureUserSaveUserSettings: {
		ID:       procedureUserSaveUserSettings,
		Kind:     "procedure",
		Summary:  "Save user settings",
		AuthMode: string(rpcAuthUser),
		Mutates:  true,
		Output:   schemaOutput{JSON: true},
		Inputs: []schemaInput{
			{Name: "settings", Type: "object", Required: true, Description: "full user settings object"},
		},
	},
	procedureUserSearch: {
		ID:       procedureUserSearch,
		Kind:     "procedure",
		Summary:  "Search using the current user profile settings",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true},
		Inputs: []schemaInput{
			{Name: "query", Type: "string", Required: true, Description: "search query"},
			{Name: "indexer_id", Type: "string", Description: "optional indexer id"},
		},
	},
}

func listCommandSchemas() []schemaEntry {
	return sortedSchemaEntries(commandSchemaRegistry)
}

func listProcedureSchemas() []schemaEntry {
	return sortedSchemaEntries(procedureSchemaRegistry)
}

func lookupCommandSchema(id string) (schemaEntry, bool) {
	entry, ok := commandSchemaRegistry[strings.TrimSpace(id)]
	return entry, ok
}

func lookupProcedureSchema(id string) (schemaEntry, bool) {
	entry, ok := procedureSchemaRegistry[strings.TrimSpace(id)]
	return entry, ok
}

func schemaCommandID(commandPath string) string {
	trimmed := strings.TrimSpace(commandPath)
	if trimmed == "" || trimmed == "chilly" {
		return "chilly"
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "chilly "))
}

func sortedSchemaEntries(registry map[string]schemaEntry) []schemaEntry {
	entries := make([]schemaEntry, 0, len(registry))
	for _, entry := range registry {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
	return entries
}

func appendInputs(extra ...schemaInput) []schemaInput {
	inputs := cloneInputs(commonCommandInputs)
	inputs = append(inputs, extra...)
	return inputs
}

func cloneInputs(inputs []schemaInput) []schemaInput {
	if len(inputs) == 0 {
		return nil
	}
	cloned := make([]schemaInput, len(inputs))
	copy(cloned, inputs)
	return cloned
}

type metadataAuthMode string

const (
	rpcAuthNone metadataAuthMode = "none"
	rpcAuthUser metadataAuthMode = "user"
)
