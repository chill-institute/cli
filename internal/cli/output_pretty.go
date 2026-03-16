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

		enabledStatus := "disabled"
		if enabled, ok := indexer["enabled"].(bool); ok && enabled {
			enabledStatus = "enabled"
		}

		line := fmt.Sprintf("%d. %s", index+1, name)
		if id := firstString(indexer, "id"); id != "" && id != name {
			line = fmt.Sprintf("%s [%s]", line, id)
		}
		lines = append(lines, line)
		lines = append(lines, fmt.Sprintf("   Enabled: %s", enabledStatus))
		if status := prettyIndexerStatus(firstString(indexer, "status")); status != "" {
			lines = append(lines, fmt.Sprintf("   Indexer Status: %s", status))
		}
	}
	if len(indexers) > maxPrettyListItems {
		lines = append(lines, fmt.Sprintf("... %d more indexers omitted. Use --output json for the full payload.", len(indexers)-maxPrettyListItems))
	}

	return strings.Join(lines, "\n"), true, nil
}

func prettyIndexerStatus(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	trimmed = strings.TrimPrefix(trimmed, "INDEXER_STATUS_")
	if trimmed == "" {
		return ""
	}

	return strings.ToLower(trimmed)
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

func renderDownloadFolderPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}
	folder, ok := payload["folder"].(map[string]any)
	if !ok {
		return "", false, nil
	}

	lines := []string{"Download Folder"}
	lines = append(lines, prettyUserFileLines(folder)...)
	return strings.Join(lines, "\n"), true, nil
}

func renderFolderPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	parent, ok := payload["parent"].(map[string]any)
	if !ok {
		return "", false, nil
	}
	files, ok := payload["files"].([]any)
	if !ok {
		return "", false, nil
	}

	lines := []string{"Folder"}
	lines = append(lines, prettyUserFileLines(parent)...)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Children: %d", len(files)))

	for index, item := range files[:min(len(files), maxPrettyListItems)] {
		file, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := firstString(file, "name")
		if name == "" {
			name = fmt.Sprintf("Item %d", index+1)
		}
		fileType := firstString(file, "fileType", "file_type")
		if fileType == "" {
			fileType = "unknown"
		}
		lines = append(lines, fmt.Sprintf("%d. %s [%s]", index+1, name, fileType))
		if id := firstString(file, "id"); id != "" {
			lines = append(lines, fmt.Sprintf("   ID: %s", id))
		}
	}
	if len(files) > maxPrettyListItems {
		lines = append(lines, fmt.Sprintf("... %d more items omitted. Use --output json for the full payload.", len(files)-maxPrettyListItems))
	}

	return strings.Join(lines, "\n"), true, nil
}

func renderTransferPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	transfer := payload
	if nested, ok := payload["transfer"].(map[string]any); ok {
		transfer = nested
	}
	if len(transfer) == 0 {
		return "", false, nil
	}

	lines := []string{"Transfer"}
	if status, ok := stringValue(payload, "status"); ok && payload["transfer"] != nil {
		lines = append(lines, fmt.Sprintf("Request Status: %s", status))
	}
	if id := prettyValue(firstPresent(transfer, "id")); id != "" && id != "<nil>" {
		lines = append(lines, fmt.Sprintf("ID: %s", id))
	}
	if name, ok := stringValue(transfer, "name"); ok {
		lines = append(lines, fmt.Sprintf("Name: %s", name))
	}
	if status := prettyValue(firstPresent(transfer, "status")); status != "" && status != "<nil>" {
		lines = append(lines, fmt.Sprintf("Status: %s", status))
	}
	if progress := prettyValue(firstPresent(transfer, "percentDone", "percent_done")); progress != "" && progress != "<nil>" {
		lines = append(lines, fmt.Sprintf("Progress: %s%%", progress))
	}
	if finished, ok := transfer["isFinished"].(bool); ok {
		lines = append(lines, fmt.Sprintf("Finished: %t", finished))
	} else if finished, ok := transfer["is_finished"].(bool); ok {
		lines = append(lines, fmt.Sprintf("Finished: %t", finished))
	}
	if message := firstString(transfer, "statusMessage", "status_message"); message != "" {
		lines = append(lines, fmt.Sprintf("Message: %s", message))
	}
	if message := firstString(transfer, "errorMessage", "error_message"); message != "" {
		lines = append(lines, fmt.Sprintf("Error: %s", message))
	}
	if fileURL := firstString(transfer, "fileUrl", "file_url"); fileURL != "" {
		lines = append(lines, fmt.Sprintf("File URL: %s", fileURL))
	}
	if fileID := prettyValue(firstPresent(transfer, "fileId", "file_id")); fileID != "" && fileID != "<nil>" {
		lines = append(lines, fmt.Sprintf("File ID: %s", fileID))
	}
	if eta := prettyValue(firstPresent(transfer, "estimatedTimeSeconds", "estimated_time_seconds")); eta != "" && eta != "<nil>" && eta != "0" {
		lines = append(lines, fmt.Sprintf("ETA Seconds: %s", eta))
	}

	return strings.Join(lines, "\n"), true, nil
}

func renderDoctorPretty(value any) (string, bool, error) {
	payload, ok := value.(map[string]any)
	if !ok {
		return "", false, nil
	}

	lines := []string{"Doctor"}
	lines = appendDoctorLine(lines, "Status", payload, "status")

	if build, ok := payload["build"].(map[string]any); ok {
		lines = append(lines, "")
		lines = append(lines, "Build")
		lines = appendDoctorLine(lines, "Version", build, "version")
		lines = appendDoctorLine(lines, "Commit", build, "commit")
		lines = appendDoctorLine(lines, "Build Date", build, "build_date")
		lines = appendDoctorLine(lines, "Dev Build", build, "dev")
	}

	if config, ok := payload["config"].(map[string]any); ok {
		lines = append(lines, "")
		lines = append(lines, "Config")
		lines = appendDoctorLine(lines, "Profile", config, "profile")
		lines = appendDoctorLine(lines, "Path", config, "path")
		lines = appendDoctorLine(lines, "Exists", config, "exists")
	}

	if api, ok := payload["api"].(map[string]any); ok {
		lines = append(lines, "")
		lines = append(lines, "API")
		lines = appendDoctorLine(lines, "Base URL", api, "base_url")
	}

	if auth, ok := payload["auth"].(map[string]any); ok {
		lines = append(lines, "")
		lines = append(lines, "Auth")
		lines = appendDoctorLine(lines, "Configured", auth, "configured")
		lines = appendDoctorLine(lines, "Status", auth, "status")
		lines = appendDoctorLine(lines, "Request ID", auth, "request_id")
		lines = appendDoctorLine(lines, "Code", auth, "code")
		lines = appendDoctorLine(lines, "Message", auth, "message")

		if user, ok := auth["user"].(map[string]any); ok {
			lines = appendDoctorLine(lines, "Username", user, "username")
			lines = appendDoctorLine(lines, "Email", user, "email")
			lines = appendDoctorLine(lines, "User ID", user, "userId", "user_id")
			lines = appendDoctorLine(lines, "Plan", user, "plan")
		}
	}

	return strings.Join(lines, "\n"), true, nil
}

func appendIfString(lines []string, label string, payload map[string]any, keys ...string) []string {
	if value := firstString(payload, keys...); value != "" {
		return append(lines, fmt.Sprintf("%s: %s", label, value))
	}
	return lines
}

func appendDoctorLine(lines []string, label string, payload map[string]any, keys ...string) []string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			lines = append(lines, fmt.Sprintf("%s: %s", label, prettyValue(value)))
			return lines
		}
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

func prettyUserFileLines(file map[string]any) []string {
	lines := make([]string, 0, 5)
	if name := firstString(file, "name"); name != "" {
		lines = append(lines, fmt.Sprintf("Name: %s", name))
	}
	if id := firstString(file, "id"); id != "" {
		lines = append(lines, fmt.Sprintf("ID: %s", id))
	}
	if fileType := firstString(file, "fileType", "file_type"); fileType != "" {
		lines = append(lines, fmt.Sprintf("Type: %s", fileType))
	}
	if createdAt := firstString(file, "createdAt", "created_at"); createdAt != "" {
		lines = append(lines, fmt.Sprintf("Created: %s", createdAt))
	}
	if shared, ok := file["isShared"].(bool); ok {
		lines = append(lines, fmt.Sprintf("Shared: %t", shared))
	} else if shared, ok := file["is_shared"].(bool); ok {
		lines = append(lines, fmt.Sprintf("Shared: %t", shared))
	}
	return lines
}

func firstPresent(payload map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			return value
		}
	}
	return nil
}

func min(left int, right int) int {
	if left < right {
		return left
	}
	return right
}
