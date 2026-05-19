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

type schemaInputMode struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs"`
	Required    []string `json:"required"`
}

type schemaOutput struct {
	JSON  bool   `json:"json"`
	Human bool   `json:"human"`
	Type  string `json:"type,omitempty"`
}

type schemaEntry struct {
	ID              string            `json:"id"`
	Kind            string            `json:"kind"`
	Summary         string            `json:"summary"`
	AliasFor        string            `json:"alias_for,omitempty"`
	AuthMode        string            `json:"auth_mode"`
	Mutates         bool              `json:"mutates"`
	SupportsDryRun  bool              `json:"supports_dry_run"`
	SupportsFields  bool              `json:"supports_fields"`
	LinkedProcedure string            `json:"linked_procedure,omitempty"`
	Inputs          []schemaInput     `json:"inputs,omitempty"`
	InputModes      []schemaInputMode `json:"input_modes,omitempty"`
	Output          schemaOutput      `json:"output"`
}

type schemaField struct {
	Name        string `json:"name"`
	JSONName    string `json:"json_name"`
	Type        string `json:"type"`
	Repeated    bool   `json:"repeated,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
	Description string `json:"description,omitempty"`
}

type schemaType struct {
	ID      string        `json:"id"`
	Kind    string        `json:"kind"`
	Summary string        `json:"summary"`
	Fields  []schemaField `json:"fields"`
}

const (
	typeSearchResponse = "chill.v4.SearchResponse"
	typeSearchResult   = "chill.v4.SearchResult"
	typeReleaseInfo    = "chill.v4.ReleaseInfo"
)

var commonCommandInputs = []schemaInput{
	{Name: "api-url", Type: "string", Description: "override API base URL"},
	{Name: "config", Type: "string", Description: "config file path"},
	{Name: "profile", Type: "string", Description: "config profile to use"},
	{Name: "output", Type: "string", Description: "output mode: pretty|json|ndjson"},
	{Name: "describe", Type: "boolean", Description: "print command metadata and exit"},
}

var commandSchemaRegistry = map[string]schemaEntry{
	"chilly": {
		ID:       "chilly",
		Kind:     "command",
		Summary:  "chill.institute CLI for humans and agents",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"add-transfer": {
		ID:              "add-transfer",
		Kind:            "command",
		Summary:         "Add a transfer through chill.institute",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		SupportsDryRun:  true,
		LinkedProcedure: procedureUserAddTransfer,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.AddTransferResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "url", Type: "string", Description: "magnet or URL to add as transfer; use either --url or --json"},
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, or @- to read it from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the request without executing it"},
		),
		InputModes: addTransferInputModes(),
	},
	"get-transfer": {
		ID:              "get-transfer",
		Kind:            "command",
		Summary:         "Show one transfer through chill.institute",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTransfer,
		Output:          schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "id", Type: "integer", Required: true, Description: "transfer id"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"auth": {
		ID:       "auth",
		Kind:     "command",
		Summary:  "Sign in and manage local auth tokens",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"auth login": {
		ID:             "auth login",
		Kind:           "command",
		Summary:        "Sign in through the hosted web token flow or store a setup token",
		AuthMode:       string(rpcAuthNone),
		Mutates:        true,
		SupportsDryRun: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "token", Type: "string", Description: "setup token to store directly"},
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, or @- to read it from stdin"},
			schemaInput{Name: "no-browser", Type: "boolean", Description: "with --local-browser, print the localhost callback URL instead of opening a browser automatically"},
			schemaInput{Name: "local-browser", Type: "boolean", Description: "use the localhost callback flow instead of the hosted web token flow"},
			schemaInput{Name: "skip-verify", Type: "boolean", Description: "skip token verification call"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "preview the auth action without verifying or saving credentials"},
		),
	},
	"auth logout": {
		ID:             "auth logout",
		Kind:           "command",
		Summary:        "Clear the stored auth token from local config",
		AuthMode:       string(rpcAuthNone),
		Mutates:        true,
		SupportsDryRun: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, or @- to read it from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the local config change without saving it"},
		),
	},
	"doctor": {
		ID:             "doctor",
		Kind:           "command",
		Summary:        "Inspect CLI health, config, and auth status",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "offline", Type: "boolean", Description: "skip auth verification and only inspect local state"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"completion": {
		ID:       "completion",
		Kind:     "command",
		Summary:  "Generate shell completion scripts",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: false, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "shell", Type: "string", Required: true, Description: "shell name: bash, zsh, fish, or powershell"},
		),
	},
	"movies": {
		ID:              "movies",
		Kind:            "command",
		Summary:         "List movies using your profile settings",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetMovies,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetMoviesResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"tv-shows": {
		ID:              "tv-shows",
		Kind:            "command",
		Summary:         "List TV shows using your profile settings",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShows,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowsResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"tv-shows detail": {
		ID:              "tv-shows detail",
		Kind:            "command",
		Summary:         "Show TV show detail by IMDb id",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowDetail,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowDetailResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"tv-shows season": {
		ID:              "tv-shows season",
		Kind:            "command",
		Summary:         "Show one TV show season by IMDb id",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowSeason,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowSeasonResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "season-number", Type: "integer", Required: true, Description: "season number"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"tv-shows episode-download": {
		ID:              "tv-shows episode-download",
		Kind:            "command",
		Summary:         "Find one TV episode download by IMDb id",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowEpisodeDownload,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowEpisodeDownloadResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "season-number", Type: "integer", Required: true, Description: "season number"},
			schemaInput{Name: "episode-number", Type: "integer", Required: true, Description: "episode number"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"tv-shows season-downloads": {
		ID:              "tv-shows season-downloads",
		Kind:            "command",
		Summary:         "Find season and episode downloads for one TV season by IMDb id",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowSeasonDownloads,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowSeasonDownloadsResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "season-number", Type: "integer", Required: true, Description: "season number"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"schema": {
		ID:             "schema",
		Kind:           "command",
		Summary:        "Inspect command and procedure contracts",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"schema command": {
		ID:             "schema command",
		Kind:           "command",
		Summary:        "Show metadata for one CLI command",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "name", Type: "string", Required: true, Description: "command id such as search or settings get"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"schema procedure": {
		ID:             "schema procedure",
		Kind:           "command",
		Summary:        "Show metadata for one backend procedure",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "procedure", Type: "string", Required: true, Description: "procedure id such as chill.v4.UserService/Search"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"schema type": {
		ID:             "schema type",
		Kind:           "command",
		Summary:        "Show metadata for one output type",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true, Type: "schema.Type"},
		Inputs: appendInputs(
			schemaInput{Name: "name", Type: "string", Required: true, Description: "type id such as chill.v4.ReleaseInfo"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"search": {
		ID:              "search",
		Kind:            "command",
		Summary:         "Search using your saved profile settings",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserSearch,
		Output:          schemaOutput{JSON: true, Human: true, Type: typeSearchResponse},
		Inputs: appendInputs(
			schemaInput{Name: "query", Type: "string", Required: true, Description: "search query"},
			schemaInput{Name: "indexer-id", Type: "string", Description: "optional indexer id; prefer one indexer at a time for agent workflows"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"self-update": {
		ID:             "self-update",
		Kind:           "command",
		Summary:        "Check for or install released CLI updates",
		AuthMode:       string(rpcAuthNone),
		Mutates:        true,
		SupportsDryRun: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "check", Type: "boolean", Description: "check for a newer release without installing it"},
			schemaInput{Name: "version", Type: "string", Description: "specific release tag to install"},
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, or @- to read it from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "preview update resolution without replacing the current executable"},
		),
	},
	"settings": {
		ID:       "settings",
		Kind:     "command",
		Summary:  "Manage local CLI config",
		AuthMode: string(rpcAuthNone),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"settings get": {
		ID:             "settings get",
		Kind:           "command",
		Summary:        "Show one local CLI setting",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "key", Type: "string", Required: true, Description: "settings key such as api-base-url"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"settings path": {
		ID:             "settings path",
		Kind:           "command",
		Summary:        "Show local config file path",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"settings set": {
		ID:             "settings set",
		Kind:           "command",
		Summary:        "Set one local CLI setting",
		AuthMode:       string(rpcAuthNone),
		Mutates:        true,
		SupportsDryRun: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "key", Type: "string", Description: "settings key such as api-base-url; use with value or use --json"},
			schemaInput{Name: "value", Type: "string", Description: "next value for the selected settings key; use with key or use --json"},
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, or @- to read it from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the local config change without saving it"},
		),
		InputModes: settingsSetInputModes(),
	},
	"settings show": {
		ID:             "settings show",
		Kind:           "command",
		Summary:        "Show local CLI config (auth token redacted)",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user": {
		ID:       "user",
		Kind:     "command",
		Summary:  "Run user account commands through chill.institute",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user indexers": {
		ID:              "user indexers",
		Kind:            "command",
		Summary:         "List user indexers",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetIndexers,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserGetIndexersResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user download-folder": {
		ID:              "user download-folder",
		Kind:            "command",
		Summary:         "Show your current download folder",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetDownloadFolder,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetDownloadFolderResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user download-folder set": {
		ID:              "user download-folder set",
		Kind:            "command",
		Summary:         "Set the current download folder",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		SupportsDryRun:  true,
		LinkedProcedure: procedureUserSaveUserSettings,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserSettings"},
		Inputs: appendInputs(
			schemaInput{Name: "id", Type: "integer", Required: true, Description: "folder id"},
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, bare settings object JSON, or @- to read from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the request or patch without executing it"},
		),
	},
	"user download-folder clear": {
		ID:              "user download-folder clear",
		Kind:            "command",
		Summary:         "Clear the current download folder setting",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		SupportsDryRun:  true,
		LinkedProcedure: procedureUserSaveUserSettings,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserSettings"},
		Inputs: appendInputs(
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, bare settings object JSON, or @- to read from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the request or patch without executing it"},
		),
	},
	"user folder": {
		ID:       "user folder",
		Kind:     "command",
		Summary:  "Inspect folders",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user folder get": {
		ID:              "user folder get",
		Kind:            "command",
		Summary:         "Get one folder and its children",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetFolder,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetFolderResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "id", Type: "integer", Required: true, Description: "folder id"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user profile": {
		ID:              "user profile",
		Kind:            "command",
		Summary:         "Show authenticated profile (alias for whoami)",
		AliasFor:        "whoami",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetUserProfile,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserProfile"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user search": {
		ID:              "user search",
		Kind:            "command",
		Summary:         "Search using your saved profile settings",
		AliasFor:        "search",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserSearch,
		Output:          schemaOutput{JSON: true, Human: true, Type: typeSearchResponse},
		Inputs: appendInputs(
			schemaInput{Name: "query", Type: "string", Required: true, Description: "search query"},
			schemaInput{Name: "indexer-id", Type: "string", Description: "optional indexer id"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user settings": {
		ID:       "user settings",
		Kind:     "command",
		Summary:  "Read and update hosted user settings",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user settings get": {
		ID:              "user settings get",
		Kind:            "command",
		Summary:         "Show hosted user settings",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetUserSettings,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserSettings"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user settings set": {
		ID:              "user settings set",
		Kind:            "command",
		Summary:         "Save full user settings JSON payload or patch one supported field",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		SupportsDryRun:  true,
		LinkedProcedure: procedureUserSaveUserSettings,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserSettings"},
		Inputs: append(appendInputs(
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, bare settings object JSON, or @- to read from stdin"},
			schemaInput{Name: "field", Type: "string", Description: "supported settings field to patch"},
			schemaInput{Name: "value", Type: "string", Description: "normalized patch value for the selected field"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the request or patch without executing it"},
		), supportedUserSettingsPatchInputs()...),
	},
	"user movies": {
		ID:              "user movies",
		Kind:            "command",
		Summary:         "List movies using your profile settings",
		AliasFor:        "movies",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetMovies,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetMoviesResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user tv-shows": {
		ID:              "user tv-shows",
		Kind:            "command",
		Summary:         "List TV shows using your profile settings",
		AliasFor:        "tv-shows",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShows,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowsResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user tv-shows detail": {
		ID:              "user tv-shows detail",
		Kind:            "command",
		Summary:         "Show TV show detail by IMDb id",
		AliasFor:        "tv-shows detail",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowDetail,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowDetailResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user tv-shows season": {
		ID:              "user tv-shows season",
		Kind:            "command",
		Summary:         "Show one TV show season by IMDb id",
		AliasFor:        "tv-shows season",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowSeason,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowSeasonResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "season-number", Type: "integer", Required: true, Description: "season number"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user tv-shows episode-download": {
		ID:              "user tv-shows episode-download",
		Kind:            "command",
		Summary:         "Find one TV episode download by IMDb id",
		AliasFor:        "tv-shows episode-download",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowEpisodeDownload,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowEpisodeDownloadResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "season-number", Type: "integer", Required: true, Description: "season number"},
			schemaInput{Name: "episode-number", Type: "integer", Required: true, Description: "episode number"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user tv-shows season-downloads": {
		ID:              "user tv-shows season-downloads",
		Kind:            "command",
		Summary:         "Find season and episode downloads for one TV season by IMDb id",
		AliasFor:        "tv-shows season-downloads",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTVShowSeasonDownloads,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTVShowSeasonDownloadsResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "imdb-id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			schemaInput{Name: "season-number", Type: "integer", Required: true, Description: "season number"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"user transfer": {
		ID:       "user transfer",
		Kind:     "command",
		Summary:  "Transfer commands",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Human: true},
		Inputs:   cloneInputs(commonCommandInputs),
	},
	"user transfer add": {
		ID:              "user transfer add",
		Kind:            "command",
		Summary:         "Add a transfer through chill.institute",
		AliasFor:        "add-transfer",
		AuthMode:        string(rpcAuthUser),
		Mutates:         true,
		SupportsDryRun:  true,
		LinkedProcedure: procedureUserAddTransfer,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.AddTransferResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "url", Type: "string", Description: "magnet or URL; use either --url or --json"},
			schemaInput{Name: "json", Type: "string", Description: "raw JSON request body, or @- to read it from stdin"},
			schemaInput{Name: "dry-run", Type: "boolean", Description: "validate input and print the request without executing it"},
		),
		InputModes: addTransferInputModes(),
	},
	"user transfer get": {
		ID:              "user transfer get",
		Kind:            "command",
		Summary:         "Show one transfer through chill.institute",
		AliasFor:        "get-transfer",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetTransfer,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.GetTransferResponse"},
		Inputs: appendInputs(
			schemaInput{Name: "id", Type: "integer", Required: true, Description: "transfer id"},
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"whoami": {
		ID:              "whoami",
		Kind:            "command",
		Summary:         "Show authenticated user profile",
		AuthMode:        string(rpcAuthUser),
		SupportsFields:  true,
		LinkedProcedure: procedureUserGetUserProfile,
		Output:          schemaOutput{JSON: true, Human: true, Type: "chill.v4.UserProfile"},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
	"version": {
		ID:             "version",
		Kind:           "command",
		Summary:        "Show CLI build info",
		AuthMode:       string(rpcAuthNone),
		SupportsFields: true,
		Output:         schemaOutput{JSON: true, Human: true},
		Inputs: appendInputs(
			schemaInput{Name: "fields", Type: "string", Description: "comma-separated field paths to include in the output"},
		),
	},
}

var procedureSchemaRegistry = map[string]schemaEntry{
	procedureUserAddTransfer: {
		ID:             procedureUserAddTransfer,
		Kind:           "procedure",
		Summary:        "Add transfer to put.io",
		AuthMode:       string(rpcAuthUser),
		Mutates:        true,
		SupportsDryRun: true,
		Output:         schemaOutput{JSON: true, Type: "chill.v4.AddTransferResponse"},
		Inputs: []schemaInput{
			{Name: "url", Type: "string", Required: true, Description: "magnet or URL to add as transfer"},
		},
	},
	procedureUserGetIndexers: {
		ID:       procedureUserGetIndexers,
		Kind:     "procedure",
		Summary:  "List user indexers",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.UserGetIndexersResponse"},
	},
	procedureUserGetTransfer: {
		ID:       procedureUserGetTransfer,
		Kind:     "procedure",
		Summary:  "Fetch one transfer",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetTransferResponse"},
		Inputs: []schemaInput{
			{Name: "id", Type: "integer", Required: true, Description: "transfer id"},
		},
	},
	procedureUserGetDownloadFolder: {
		ID:       procedureUserGetDownloadFolder,
		Kind:     "procedure",
		Summary:  "Fetch the current download folder",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetDownloadFolderResponse"},
	},
	procedureUserGetFolder: {
		ID:       procedureUserGetFolder,
		Kind:     "procedure",
		Summary:  "Fetch one folder and its children",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetFolderResponse"},
		Inputs: []schemaInput{
			{Name: "id", Type: "integer", Required: true, Description: "folder id"},
		},
	},
	procedureUserGetMovies: {
		ID:       procedureUserGetMovies,
		Kind:     "procedure",
		Summary:  "List movies for the current user",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetMoviesResponse"},
	},
	procedureUserGetTVShows: {
		ID:       procedureUserGetTVShows,
		Kind:     "procedure",
		Summary:  "List TV shows for the current user",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetTVShowsResponse"},
	},
	procedureUserGetTVShowDetail: {
		ID:       procedureUserGetTVShowDetail,
		Kind:     "procedure",
		Summary:  "Fetch TV show detail by IMDb id",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetTVShowDetailResponse"},
		Inputs: []schemaInput{
			{Name: "imdb_id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
		},
	},
	procedureUserGetTVShowSeason: {
		ID:       procedureUserGetTVShowSeason,
		Kind:     "procedure",
		Summary:  "Fetch one TV show season by IMDb id",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetTVShowSeasonResponse"},
		Inputs: []schemaInput{
			{Name: "imdb_id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			{Name: "season_number", Type: "integer", Required: true, Description: "season number"},
		},
	},
	procedureUserGetTVShowEpisodeDownload: {
		ID:       procedureUserGetTVShowEpisodeDownload,
		Kind:     "procedure",
		Summary:  "Find one TV episode download by IMDb id",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetTVShowEpisodeDownloadResponse"},
		Inputs: []schemaInput{
			{Name: "imdb_id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			{Name: "season_number", Type: "integer", Required: true, Description: "season number"},
			{Name: "episode_number", Type: "integer", Required: true, Description: "episode number"},
		},
	},
	procedureUserGetTVShowSeasonDownloads: {
		ID:       procedureUserGetTVShowSeasonDownloads,
		Kind:     "procedure",
		Summary:  "Find season and episode downloads for one TV season by IMDb id",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.GetTVShowSeasonDownloadsResponse"},
		Inputs: []schemaInput{
			{Name: "imdb_id", Type: "string", Required: true, Description: "IMDb id such as tt0944947"},
			{Name: "season_number", Type: "integer", Required: true, Description: "season number"},
		},
	},
	procedureUserGetUserProfile: {
		ID:       procedureUserGetUserProfile,
		Kind:     "procedure",
		Summary:  "Fetch the authenticated user profile",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.UserProfile"},
	},
	procedureUserGetUserSettings: {
		ID:       procedureUserGetUserSettings,
		Kind:     "procedure",
		Summary:  "Fetch user settings",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: "chill.v4.UserSettings"},
	},
	procedureUserSaveUserSettings: {
		ID:             procedureUserSaveUserSettings,
		Kind:           "procedure",
		Summary:        "Save user settings",
		AuthMode:       string(rpcAuthUser),
		Mutates:        true,
		SupportsDryRun: true,
		Output:         schemaOutput{JSON: true, Type: "chill.v4.UserSettings"},
		Inputs: []schemaInput{
			{Name: "settings", Type: "object", Required: true, Description: "full user settings object"},
		},
	},
	procedureUserSearch: {
		ID:       procedureUserSearch,
		Kind:     "procedure",
		Summary:  "Search using the current user profile settings",
		AuthMode: string(rpcAuthUser),
		Output:   schemaOutput{JSON: true, Type: typeSearchResponse},
		Inputs: []schemaInput{
			{Name: "query", Type: "string", Required: true, Description: "search query"},
			{Name: "indexer_id", Type: "string", Description: "optional indexer id"},
		},
	},
}

var typeSchemaRegistry = map[string]schemaType{
	typeReleaseInfo: {
		ID:      typeReleaseInfo,
		Kind:    "type",
		Summary: "Parsed release metadata attached to search results when the API can infer it from torrent names",
		Fields: []schemaField{
			{Name: "title", JSONName: "title", Type: "string"},
			{Name: "year", JSONName: "year", Type: "integer", Optional: true},
			{Name: "season", JSONName: "season", Type: "integer", Optional: true},
			{Name: "episode", JSONName: "episode", Type: "integer", Optional: true},
			{Name: "episode_end", JSONName: "episodeEnd", Type: "integer", Optional: true},
			{Name: "part", JSONName: "part", Type: "integer", Optional: true},
			{Name: "resolution", JSONName: "resolution", Type: "string"},
			{Name: "quality", JSONName: "quality", Type: "string"},
			{Name: "source", JSONName: "source", Type: "string"},
			{Name: "codec", JSONName: "codec", Type: "string"},
			{Name: "hdr", JSONName: "hdr", Type: "string"},
			{Name: "audio", JSONName: "audio", Type: "string"},
			{Name: "group", JSONName: "group", Type: "string"},
			{Name: "container", JSONName: "container", Type: "string"},
			{Name: "language", JSONName: "language", Type: "string"},
			{Name: "region", JSONName: "region", Type: "string"},
			{Name: "size", JSONName: "size", Type: "string"},
			{Name: "bit_depth", JSONName: "bitDepth", Type: "string"},
			{Name: "edition", JSONName: "edition", Type: "string"},
			{Name: "extended", JSONName: "extended", Type: "boolean"},
			{Name: "hardcoded", JSONName: "hardcoded", Type: "boolean"},
			{Name: "proper", JSONName: "proper", Type: "boolean"},
			{Name: "repack", JSONName: "repack", Type: "boolean"},
			{Name: "remastered", JSONName: "remastered", Type: "boolean"},
			{Name: "complete", JSONName: "complete", Type: "boolean"},
			{Name: "three_d", JSONName: "threeD", Type: "boolean"},
			{Name: "imax", JSONName: "imax", Type: "boolean"},
			{Name: "unrated", JSONName: "unrated", Type: "boolean"},
			{Name: "widescreen", JSONName: "widescreen", Type: "boolean"},
			{Name: "excess", JSONName: "excess", Type: "string"},
		},
	},
	typeSearchResponse: {
		ID:      typeSearchResponse,
		Kind:    "type",
		Summary: "Search response returned by user search procedures",
		Fields: []schemaField{
			{Name: "query", JSONName: "query", Type: "string"},
			{Name: "results", JSONName: "results", Type: typeSearchResult, Repeated: true},
		},
	},
	typeSearchResult: {
		ID:      typeSearchResult,
		Kind:    "type",
		Summary: "One torrent search result",
		Fields: []schemaField{
			{Name: "id", JSONName: "id", Type: "string"},
			{Name: "title", JSONName: "title", Type: "string"},
			{Name: "indexer", JSONName: "indexer", Type: "string"},
			{Name: "link", JSONName: "link", Type: "string"},
			{Name: "imdb_id", JSONName: "imdbId", Type: "string", Optional: true},
			{Name: "peers", JSONName: "peers", Type: "integer"},
			{Name: "seeders", JSONName: "seeders", Type: "integer"},
			{Name: "size", JSONName: "size", Type: "integer"},
			{Name: "source", JSONName: "source", Type: "string"},
			{Name: "uploaded_at", JSONName: "uploadedAt", Type: "string"},
			{Name: "release_info", JSONName: "releaseInfo", Type: typeReleaseInfo, Optional: true},
		},
	},
}

func listCommandSchemas() []schemaEntry {
	return sortedSchemaEntries(commandSchemaRegistry)
}

func listProcedureSchemas() []schemaEntry {
	return sortedSchemaEntries(procedureSchemaRegistry)
}

func listTypeSchemas() []schemaType {
	entries := make([]schemaType, 0, len(typeSchemaRegistry))
	for _, entry := range typeSchemaRegistry {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
	return entries
}

func lookupCommandSchema(id string) (schemaEntry, bool) {
	entry, ok := commandSchemaRegistry[strings.TrimSpace(id)]
	return entry, ok
}

func lookupProcedureSchema(id string) (schemaEntry, bool) {
	entry, ok := procedureSchemaRegistry[strings.TrimSpace(id)]
	return entry, ok
}

func lookupTypeSchema(id string) (schemaType, bool) {
	entry, ok := typeSchemaRegistry[strings.TrimSpace(id)]
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

func addTransferInputModes() []schemaInputMode {
	return []schemaInputMode{
		schemaInputModeForRequiredInputs(
			"url",
			"Build the request from --url.",
			"url",
		),
		schemaInputModeForRequiredInputs(
			"json",
			"Use a raw JSON request body from --json.",
			"json",
		),
	}
}

func settingsSetInputModes() []schemaInputMode {
	return []schemaInputMode{
		schemaInputModeForRequiredInputs(
			"key-value",
			"Set one setting from positional key and value arguments.",
			"key",
			"value",
		),
		schemaInputModeForRequiredInputs(
			"json",
			"Set one setting from a raw JSON request body.",
			"json",
		),
	}
}

func schemaInputModeForRequiredInputs(name string, description string, inputs ...string) schemaInputMode {
	return schemaInputMode{
		Name:        name,
		Description: description,
		Inputs:      cloneStrings(inputs),
		Required:    cloneStrings(inputs),
	}
}

func cloneInputs(inputs []schemaInput) []schemaInput {
	if len(inputs) == 0 {
		return nil
	}
	cloned := make([]schemaInput, len(inputs))
	copy(cloned, inputs)
	return cloned
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

type metadataAuthMode string

const (
	rpcAuthNone metadataAuthMode = "none"
	rpcAuthUser metadataAuthMode = "user"
)
