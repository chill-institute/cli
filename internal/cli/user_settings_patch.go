package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type userSettingsPatch struct {
	Field string `json:"field"`
	Value any    `json:"value"`
}

var userSettingsPatchSpecs = []struct {
	field       string
	aliases     []string
	valueType   string
	description string
	normalize   func(string) (any, error)
}{
	{
		field:       "showTopMovies",
		aliases:     []string{"show-top-movies"},
		valueType:   "boolean",
		description: "whether top movies should be shown",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "showPrettyNamesForTopMovies",
		aliases:     []string{"show-pretty-names-for-top-movies"},
		valueType:   "boolean",
		description: "whether top movies should use provider pretty names",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "filterNastyResults",
		aliases:     []string{"filter-nasty-results"},
		valueType:   "boolean",
		description: "whether nasty results should be filtered",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "filterResultsWithNoSeeders",
		aliases:     []string{"filter-results-with-no-seeders"},
		valueType:   "boolean",
		description: "whether results with no seeders should be filtered",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "rememberQuickFilters",
		aliases:     []string{"remember-quick-filters"},
		valueType:   "boolean",
		description: "whether quick filters should be remembered",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "downloadFolderId",
		aliases:     []string{"download-folder-id"},
		valueType:   "integer-or-null",
		description: "download folder id, or null to clear it",
		normalize:   normalizeNullableInt64Value,
	},
	{
		field:       "sortBy",
		aliases:     []string{"sort-by"},
		valueType:   "enum",
		description: "one of: title, seeders, size, uploaded-at, source",
		normalize: normalizeEnumValue(map[string]string{
			"title":       "SORT_BY_TITLE",
			"seeders":     "SORT_BY_SEEDERS",
			"size":        "SORT_BY_SIZE",
			"uploaded-at": "SORT_BY_UPLOADED_AT",
			"uploaded_at": "SORT_BY_UPLOADED_AT",
			"source":      "SORT_BY_SOURCE",
		}),
	},
	{
		field:       "sortDirection",
		aliases:     []string{"sort-direction"},
		valueType:   "enum",
		description: "one of: asc, desc",
		normalize: normalizeEnumValue(map[string]string{
			"asc":  "SORT_DIRECTION_ASC",
			"desc": "SORT_DIRECTION_DESC",
		}),
	},
	{
		field:       "searchResultDisplayBehavior",
		aliases:     []string{"search-result-display-behavior"},
		valueType:   "enum",
		description: "one of: all, fastest",
		normalize: normalizeEnumValue(map[string]string{
			"all":     "SEARCH_RESULT_DISPLAY_BEHAVIOR_ALL",
			"fastest": "SEARCH_RESULT_DISPLAY_BEHAVIOR_FASTEST",
		}),
	},
	{
		field:       "searchResultTitleBehavior",
		aliases:     []string{"search-result-title-behavior"},
		valueType:   "enum",
		description: "one of: link, text",
		normalize: normalizeEnumValue(map[string]string{
			"link": "SEARCH_RESULT_TITLE_BEHAVIOR_LINK",
			"text": "SEARCH_RESULT_TITLE_BEHAVIOR_TEXT",
		}),
	},
	{
		field:       "topMoviesDisplayType",
		aliases:     []string{"top-movies-display-type"},
		valueType:   "enum",
		description: "one of: compact, expanded",
		normalize: normalizeEnumValue(map[string]string{
			"compact":  "TOP_MOVIES_DISPLAY_TYPE_COMPACT",
			"expanded": "TOP_MOVIES_DISPLAY_TYPE_EXPANDED",
		}),
	},
	{
		field:       "topMoviesSource",
		aliases:     []string{"top-movies-source"},
		valueType:   "enum",
		description: "one of: imdb-moviemeter, imdb-top-250, yts, rotten-tomatoes, trakt",
		normalize: normalizeEnumValue(map[string]string{
			"imdb-moviemeter": "TOP_MOVIES_SOURCE_IMDB_MOVIEMETER",
			"imdb/moviemeter": "TOP_MOVIES_SOURCE_IMDB_MOVIEMETER",
			"imdb-top-250":    "TOP_MOVIES_SOURCE_IMDB_TOP_250",
			"imdb/top-250":    "TOP_MOVIES_SOURCE_IMDB_TOP_250",
			"yts":             "TOP_MOVIES_SOURCE_YTS",
			"rotten-tomatoes": "TOP_MOVIES_SOURCE_ROTTEN_TOMATOES",
			"rotten_tomatoes": "TOP_MOVIES_SOURCE_ROTTEN_TOMATOES",
			"rottentomatoes":  "TOP_MOVIES_SOURCE_ROTTEN_TOMATOES",
			"trakt":           "TOP_MOVIES_SOURCE_TRAKT",
		}),
	},
}

func normalizeUserSettingsPatch(field string, value string) (userSettingsPatch, error) {
	normalizedField := normalizePatchFieldName(field)

	for _, spec := range userSettingsPatchSpecs {
		if normalizedField == spec.field {
			normalizedValue, err := spec.normalize(value)
			if err != nil {
				return userSettingsPatch{}, err
			}
			return userSettingsPatch{
				Field: spec.field,
				Value: normalizedValue,
			}, nil
		}
	}

	return userSettingsPatch{}, usageError("unsupported_user_settings_field", "unsupported user settings field %q", field)
}

func applyUserSettingsPatch(settings map[string]any, patch userSettingsPatch) map[string]any {
	cloned := cloneJSONObject(settings)
	cloned[patch.Field] = patch.Value
	return cloned
}

func supportedUserSettingsPatchInputs() []schemaInput {
	inputs := make([]schemaInput, 0, len(userSettingsPatchSpecs)*2)
	for _, spec := range userSettingsPatchSpecs {
		inputs = append(inputs,
			schemaInput{
				Name:        fmt.Sprintf("field:%s", spec.field),
				Type:        spec.valueType,
				Description: spec.description,
			},
		)
	}
	return inputs
}

func supportedUserSettingsPatchHelp() string {
	lines := make([]string, 0, len(userSettingsPatchSpecs)+1)
	lines = append(lines, "Supported patch fields:")
	for _, spec := range userSettingsPatchSpecs {
		lines = append(lines, fmt.Sprintf("  - %s (%s): %s", kebabCase(spec.field), spec.valueType, spec.description))
	}
	return strings.Join(lines, "\n")
}

func normalizePatchFieldName(raw string) string {
	trimmed := strings.TrimSpace(raw)
	for _, spec := range userSettingsPatchSpecs {
		if strings.EqualFold(trimmed, spec.field) {
			return spec.field
		}
		for _, alias := range spec.aliases {
			if strings.EqualFold(trimmed, alias) {
				return spec.field
			}
		}
	}
	return trimmed
}

func kebabCase(raw string) string {
	var builder strings.Builder
	for index, r := range raw {
		if index > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteByte('-')
		}
		builder.WriteRune(r)
	}
	return strings.ToLower(builder.String())
}

func normalizeBooleanValue(raw string) (any, error) {
	parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return nil, usageError("invalid_user_settings_value", "expected boolean value, got %q", raw)
	}
	return parsed, nil
}

func normalizeNullableInt64Value(raw string) (any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, usageError("invalid_user_settings_value", "expected integer or null, got empty value")
	}
	if strings.EqualFold(trimmed, "null") || strings.EqualFold(trimmed, "none") {
		return nil, nil
	}

	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return nil, usageError("invalid_user_settings_value", "expected integer or null, got %q", raw)
	}
	return strconv.FormatInt(parsed, 10), nil
}

func normalizeEnumValue(values map[string]string) func(string) (any, error) {
	return func(raw string) (any, error) {
		trimmed := strings.TrimSpace(strings.ToLower(raw))
		if normalized, ok := values[trimmed]; ok {
			return normalized, nil
		}
		return nil, usageError("invalid_user_settings_value", "unsupported value %q", raw)
	}
}

func cloneJSONObject(source map[string]any) map[string]any {
	if len(source) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		switch typed := value.(type) {
		case map[string]any:
			cloned[key] = cloneJSONObject(typed)
		case []any:
			next := make([]any, len(typed))
			copy(next, typed)
			cloned[key] = next
		default:
			cloned[key] = value
		}
	}
	return cloned
}
