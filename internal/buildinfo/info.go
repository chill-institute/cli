package buildinfo

import "strings"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

func Current() Info {
	return Info{
		Version:   normalizedValue(version, "dev"),
		Commit:    normalizedValue(commit, "unknown"),
		BuildDate: normalizedValue(date, "unknown"),
	}
}

func (info Info) IsDev() bool {
	return strings.EqualFold(strings.TrimSpace(info.Version), "dev")
}

func normalizedValue(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
