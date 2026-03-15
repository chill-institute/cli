package cli

import (
	"fmt"
	"sort"
	"strings"
)

const maxPrettyListItems = 10

func renderWhoamiPretty(value any) (string, bool, error) {
	profile, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	lines := make([]string, 0, 6)
	lines = appendIfString(lines, "Username", profile, "username")
	lines = appendIfString(lines, "Email", profile, "email")
	lines = appendIfString(lines, "User ID", profile, "userId", "user_id")
	lines = appendIfString(lines, "Plan", profile, "plan")

	if len(lines) == 0 {
		return "", false, nil
	}
	return strings.Join(lines, "\n"), true, nil
}

func renderSearchPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	results, ok := payload["results"].([]any)
	if !ok {
		return "", false, nil
	}

	lines := []string{
		fmt.Sprintf("Results: %d", len(results)),
	}
	if query, ok := stringValue(payload, "query"); ok {
		lines = append(lines, fmt.Sprintf("Query: %s", query))
	}

	if len(results) == 0 {
		lines = append(lines, "", "No results.")
		return strings.Join(lines, "\n"), true, nil
	}

	for index, item := range results[:min(len(results), maxPrettyListItems)] {
		result, ok := item.(map[string]any)
		if !ok {
			continue
		}

		title := firstString(result, "title", "name")
		if title == "" {
			title = fmt.Sprintf("Result %d", index+1)
		}

		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("%d. %s", index+1, title))
		lines = appendDetailLine(lines, "Indexer", result, "indexerName", "indexer_name")
		lines = appendDetailLine(lines, "Size", result, "size")
		lines = appendDetailLine(lines, "Seeds", result, "seeders", "seeds")
		lines = appendDetailLine(lines, "Peers", result, "peers", "leechers")
		lines = appendDetailLine(lines, "Magnet", result, "magnetLink", "magnet_link")
	}
	if len(results) > maxPrettyListItems {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("... %d more results omitted. Use --output json or --fields for the full payload.", len(results)-maxPrettyListItems))
	}

	return strings.Join(lines, "\n"), true, nil
}

func renderTopMoviesPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	movies, ok := payload["movies"].([]any)
	if !ok {
		return "", false, nil
	}

	lines := []string{
		fmt.Sprintf("Movies: %d", len(movies)),
	}
	if len(movies) == 0 {
		lines = append(lines, "", "No top movies yet.")
		return strings.Join(lines, "\n"), true, nil
	}

	for index, item := range movies[:min(len(movies), maxPrettyListItems)] {
		movie, ok := item.(map[string]any)
		if !ok {
			continue
		}

		title := firstString(movie, "title", "name")
		if title == "" {
			title = fmt.Sprintf("Movie %d", index+1)
		}

		year := firstString(movie, "year")
		if year != "" {
			title = fmt.Sprintf("%s (%s)", title, year)
		}

		lines = append(lines, fmt.Sprintf("%d. %s", index+1, title))
		lines = appendDetailLine(lines, "IMDb", movie, "imdbRating", "imdb_rating")
		lines = appendDetailLine(lines, "TMDb", movie, "tmdbRating", "tmdb_rating")
	}
	if len(movies) > maxPrettyListItems {
		lines = append(lines, fmt.Sprintf("... %d more movies omitted. Use --output json or --fields for the full payload.", len(movies)-maxPrettyListItems))
	}

	return strings.Join(lines, "\n"), true, nil
}

func renderUserIndexersPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	indexers, ok := payload["indexers"].([]any)
	if !ok {
		return "", false, nil
	}

	lines := []string{
		fmt.Sprintf("Indexers: %d", len(indexers)),
	}
	if len(indexers) == 0 {
		lines = append(lines, "", "No indexers configured.")
		return strings.Join(lines, "\n"), true, nil
	}

	for index, item := range indexers[:min(len(indexers), maxPrettyListItems)] {
		indexer, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name := firstString(indexer, "name")
		if name == "" {
			name = firstString(indexer, "id")
		}
		if name == "" {
			name = fmt.Sprintf("Indexer %d", index+1)
		}

		status := "disabled"
		if enabled, ok := indexer["enabled"].(bool); ok && enabled {
			status = "enabled"
		}

		line := fmt.Sprintf("%d. %s", index+1, name)
		if id := firstString(indexer, "id"); id != "" && id != name {
			line = fmt.Sprintf("%s [%s]", line, id)
		}
		lines = append(lines, line)
		lines = append(lines, fmt.Sprintf("   Status: %s", status))
	}
	if len(indexers) > maxPrettyListItems {
		lines = append(lines, fmt.Sprintf("... %d more indexers omitted. Use --output json for the full payload.", len(indexers)-maxPrettyListItems))
	}

	return strings.Join(lines, "\n"), true, nil
}

func renderUserSettingsPretty(value any) (string, bool, error) {
	settings, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}
	if len(settings) == 0 {
		return "User settings are empty.", true, nil
	}

	keys := make([]string, 0, len(settings))
	for key := range settings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := []string{"User Settings"}
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, prettyValue(settings[key])))
	}
	return strings.Join(lines, "\n"), true, nil
}

func appendIfString(lines []string, label string, payload map[string]any, keys ...string) []string {
	if value := firstString(payload, keys...); value != "" {
		return append(lines, fmt.Sprintf("%s: %s", label, value))
	}
	return lines
}

func appendDetailLine(lines []string, label string, payload map[string]any, keys ...string) []string {
	if value := firstString(payload, keys...); value != "" {
		return append(lines, fmt.Sprintf("   %s: %s", label, value))
	}
	return lines
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := stringValue(payload, key); ok {
			return value
		}
	}
	return ""
}

func stringValue(payload map[string]any, key string) (string, bool) {
	value, ok := payload[key]
	if !ok || value == nil {
		return "", false
	}

	switch typed := value.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		return trimmed, trimmed != ""
	case float64:
		return formatNumeric(typed), true
	case bool:
		if typed {
			return "true", true
		}
		return "false", true
	default:
		return "", false
	}
}

func formatNumeric(value float64) string {
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d", int64(value))
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

func prettyValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return typed
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case float64:
		return formatNumeric(typed)
	case []any:
		if len(typed) == 0 {
			return "[]"
		}
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, prettyValue(item))
		}
		return strings.Join(items, ", ")
	default:
		return fmt.Sprintf("%v", value)
	}
}

func min(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
