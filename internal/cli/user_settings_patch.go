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

type userSettingsPatchSpec struct {
	field       string
	aliases     []string
	path        []string
	valueType   string
	description string
	normalize   func(string) (any, error)
}

var userSettingsPatchSpecs = []userSettingsPatchSpec{
	{
		field:       "filterNastyResults",
		aliases:     []string{"filter-nasty-results", "search.filter-nasty-results"},
		path:        []string{"search", "filterNastyResults"},
		valueType:   "boolean",
		description: "whether nasty results should be filtered",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "filterResultsWithNoSeeders",
		aliases:     []string{"filter-results-with-no-seeders", "search.filter-results-with-no-seeders"},
		path:        []string{"search", "filterResultsWithNoSeeders"},
		valueType:   "boolean",
		description: "whether results with no seeders should be filtered",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "rememberQuickFilters",
		aliases:     []string{"remember-quick-filters", "search.remember-quick-filters"},
		path:        []string{"search", "rememberQuickFilters"},
		valueType:   "boolean",
		description: "whether quick filters should be remembered",
		normalize:   normalizeBooleanValue,
	},
	{
		field:       "downloadFolderId",
		aliases:     []string{"download-folder-id", "download.folder-id", "download.folderId"},
		path:        []string{"download", "folderId"},
		valueType:   "integer-or-null",
		description: "download folder id, or null to clear it",
		normalize:   normalizeNullableInt64Value,
	},
	{
		field:       "sortBy",
		aliases:     []string{"sort-by", "search.sort-by"},
		path:        []string{"search", "sortBy"},
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
		aliases:     []string{"sort-direction", "search.sort-direction"},
		path:        []string{"search", "sortDirection"},
		valueType:   "enum",
		description: "one of: asc, desc",
		normalize: normalizeEnumValue(map[string]string{
			"asc":  "SORT_DIRECTION_ASC",
			"desc": "SORT_DIRECTION_DESC",
		}),
	},
	{
		field:       "searchResultDisplayBehavior",
		aliases:     []string{"search-result-display-behavior", "search.search-result-display-behavior"},
		path:        []string{"search", "searchResultDisplayBehavior"},
		valueType:   "enum",
		description: "one of: all, fastest",
		normalize: normalizeEnumValue(map[string]string{
			"all":     "SEARCH_RESULT_DISPLAY_BEHAVIOR_ALL",
			"fastest": "SEARCH_RESULT_DISPLAY_BEHAVIOR_FASTEST",
		}),
	},
	{
		field:       "searchResultTitleBehavior",
		aliases:     []string{"search-result-title-behavior", "search.search-result-title-behavior"},
		path:        []string{"search", "searchResultTitleBehavior"},
		valueType:   "enum",
		description: "one of: link, text",
		normalize: normalizeEnumValue(map[string]string{
			"link": "SEARCH_RESULT_TITLE_BEHAVIOR_LINK",
			"text": "SEARCH_RESULT_TITLE_BEHAVIOR_TEXT",
		}),
	},
	{
		field:       "moviesSource",
		aliases:     []string{"movies-source", "catalog.movies-source"},
		path:        []string{"catalog", "moviesSource"},
		valueType:   "enum",
		description: "one of: imdb-moviemeter, imdb-top-250, yts, rotten-tomatoes, trakt",
		normalize: normalizeEnumValue(map[string]string{
			"imdb-moviemeter": "MOVIES_SOURCE_IMDB_MOVIEMETER",
			"imdb/moviemeter": "MOVIES_SOURCE_IMDB_MOVIEMETER",
			"imdb-top-250":    "MOVIES_SOURCE_IMDB_TOP_250",
			"imdb/top-250":    "MOVIES_SOURCE_IMDB_TOP_250",
			"yts":             "MOVIES_SOURCE_YTS",
			"rotten-tomatoes": "MOVIES_SOURCE_ROTTEN_TOMATOES",
			"rotten_tomatoes": "MOVIES_SOURCE_ROTTEN_TOMATOES",
			"rottentomatoes":  "MOVIES_SOURCE_ROTTEN_TOMATOES",
			"trakt":           "MOVIES_SOURCE_TRAKT",
		}),
	},
	{
		field:       "tvShowsSource",
		aliases:     []string{"tv-shows-source", "catalog.tv-shows-source"},
		path:        []string{"catalog", "tvShowsSource"},
		valueType:   "enum",
		description: "one of: netflix, hbo-max, apple-tv-plus, prime-video, disney-plus",
		normalize: normalizeEnumValue(map[string]string{
			"netflix":       "TV_SHOWS_SOURCE_NETFLIX",
			"hbo-max":       "TV_SHOWS_SOURCE_HBO_MAX",
			"hbo_max":       "TV_SHOWS_SOURCE_HBO_MAX",
			"apple-tv-plus": "TV_SHOWS_SOURCE_APPLE_TV_PLUS",
			"apple_tv_plus": "TV_SHOWS_SOURCE_APPLE_TV_PLUS",
			"prime-video":   "TV_SHOWS_SOURCE_PRIME_VIDEO",
			"prime_video":   "TV_SHOWS_SOURCE_PRIME_VIDEO",
			"disney-plus":   "TV_SHOWS_SOURCE_DISNEY_PLUS",
			"disney_plus":   "TV_SHOWS_SOURCE_DISNEY_PLUS",
		}),
	},
}

var legacyFlatUserSettingsPaths = map[string][]string{
	"codecFilters":                {"search", "codecFilters"},
	"disabledIndexerIds":          {"search", "disabledIndexerIds"},
	"downloadFolderId":            {"download", "folderId"},
	"filterNastyResults":          {"search", "filterNastyResults"},
	"filterResultsWithNoSeeders":  {"search", "filterResultsWithNoSeeders"},
	"moviesSource":                {"catalog", "moviesSource"},
	"otherFilters":                {"search", "otherFilters"},
	"rememberQuickFilters":        {"search", "rememberQuickFilters"},
	"resolutionFilters":           {"search", "resolutionFilters"},
	"searchResultDisplayBehavior": {"search", "searchResultDisplayBehavior"},
	"searchResultTitleBehavior":   {"search", "searchResultTitleBehavior"},
	"sortBy":                      {"search", "sortBy"},
	"sortDirection":               {"search", "sortDirection"},
	"tvShowsSource":               {"catalog", "tvShowsSource"},
}

var obsoleteFlatUserSettingsFields = []string{
	"cardDisplayType",
	"showMovies",
	"showTvShows",
}

func normalizeUserSettingsPatch(field string, value string) (userSettingsPatch, error) {
	spec, ok := userSettingsPatchSpecForField(field)
	if !ok {
		return userSettingsPatch{}, usageError("unsupported_user_settings_field", "unsupported user settings field %q", field)
	}

	normalizedValue, err := spec.normalize(value)
	if err != nil {
		return userSettingsPatch{}, err
	}
	return userSettingsPatch{
		Field: strings.Join(spec.path, "."),
		Value: normalizedValue,
	}, nil
}

func applyUserSettingsPatch(settings map[string]any, patch userSettingsPatch) map[string]any {
	cloned := normalizeLegacyFlatUserSettings(settings)
	setNestedJSONObjectValue(cloned, strings.Split(patch.Field, "."), patch.Value)
	return cloned
}

func supportedUserSettingsPatchInputs() []schemaInput {
	inputs := make([]schemaInput, 0, len(userSettingsPatchSpecs)*2)
	for _, spec := range userSettingsPatchSpecs {
		inputs = append(inputs,
			schemaInput{
				Name:        fmt.Sprintf("field:%s", strings.Join(spec.path, ".")),
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
	if spec, ok := userSettingsPatchSpecForField(trimmed); ok {
		return spec.field
	}
	return trimmed
}

func userSettingsPatchSpecForField(raw string) (userSettingsPatchSpec, bool) {
	trimmed := strings.TrimSpace(raw)
	for _, spec := range userSettingsPatchSpecs {
		if strings.EqualFold(trimmed, spec.field) {
			return spec, true
		}
		if strings.EqualFold(trimmed, strings.Join(spec.path, ".")) {
			return spec, true
		}
		for _, alias := range spec.aliases {
			if strings.EqualFold(trimmed, alias) {
				return spec, true
			}
		}
	}
	return userSettingsPatchSpec{}, false
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

func setNestedJSONObjectValue(target map[string]any, path []string, value any) {
	if len(path) == 0 {
		return
	}
	if len(path) == 1 {
		target[path[0]] = value
		return
	}
	next, ok := target[path[0]].(map[string]any)
	if !ok {
		next = map[string]any{}
		target[path[0]] = next
	}
	setNestedJSONObjectValue(next, path[1:], value)
}

func hasNestedJSONObjectValue(target map[string]any, path []string) bool {
	if len(path) == 0 {
		return false
	}
	if len(path) == 1 {
		_, ok := target[path[0]]
		return ok
	}
	next, ok := target[path[0]].(map[string]any)
	if !ok {
		return false
	}
	return hasNestedJSONObjectValue(next, path[1:])
}

func normalizeLegacyFlatUserSettings(settings map[string]any) map[string]any {
	cloned := cloneJSONObject(settings)
	for _, field := range obsoleteFlatUserSettingsFields {
		delete(cloned, field)
	}
	for field, path := range legacyFlatUserSettingsPaths {
		value, ok := cloned[field]
		if !ok {
			continue
		}
		delete(cloned, field)
		if !hasNestedJSONObjectValue(cloned, path) {
			setNestedJSONObjectValue(cloned, path, value)
		}
	}
	return cloned
}
