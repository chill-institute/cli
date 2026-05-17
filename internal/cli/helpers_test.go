package cli

import (
	"strings"
	"testing"
)

func TestNormalizeDownloadFolderJSONValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		value     any
		allowNull bool
		want      any
		wantErr   bool
	}{
		{name: "string", value: "42", want: "42"},
		{name: "float", value: float64(42), want: "42"},
		{name: "null allowed", value: nil, allowNull: true, want: nil},
		{name: "negative", value: float64(-1), wantErr: true},
		{name: "fraction", value: float64(1.5), wantErr: true},
		{name: "null disallowed", value: nil, wantErr: true},
		{name: "unsupported type", value: true, allowNull: true, wantErr: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeDownloadFolderJSONValue(tc.value, tc.allowNull)
			if tc.wantErr {
				if err == nil {
					t.Fatal("normalizeDownloadFolderJSONValue() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeDownloadFolderJSONValue() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalizeDownloadFolderJSONValue() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestResolveDownloadFolderRequests(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	setRequest, err := resolveDownloadFolderSetRequest(app, `{"downloadFolderId":42}`)
	if err != nil {
		t.Fatalf("resolveDownloadFolderSetRequest() error = %v", err)
	}
	if setRequest.Field != "download.folderId" || setRequest.Value != "42" {
		t.Fatalf("setRequest = %#v", setRequest)
	}

	clearRequest, err := resolveDownloadFolderClearRequest(app, `{"settings":{"download":{"folderId":null}}}`)
	if err != nil {
		t.Fatalf("resolveDownloadFolderClearRequest() error = %v", err)
	}
	if clearRequest.Field != "download.folderId" || clearRequest.Value != nil {
		t.Fatalf("clearRequest = %#v", clearRequest)
	}

	if _, err := resolveDownloadFolderRequest(app, `{}`, false); err == nil {
		t.Fatal("resolveDownloadFolderRequest() error = nil, want missing field error")
	}
	if _, err := resolveDownloadFolderClearRequest(app, `{"settings":{"downloadFolderId":42}}`); err == nil {
		t.Fatal("resolveDownloadFolderClearRequest() error = nil, want null requirement error")
	}
}

func TestDecodeUserSettingsRequestNormalizesLegacyFlatSettings(t *testing.T) {
	t.Parallel()

	app := &appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: &strings.Builder{},
		stderr: &strings.Builder{},
	}

	request, err := decodeUserSettingsRequest(app, `{"sortBy":"SORT_BY_TITLE","downloadFolderId":42,"moviesSource":"MOVIES_SOURCE_YTS"}`)
	if err != nil {
		t.Fatalf("decodeUserSettingsRequest() error = %v", err)
	}
	settings := request["settings"].(map[string]any)
	if settings["sortBy"] != nil || settings["downloadFolderId"] != nil || settings["moviesSource"] != nil {
		t.Fatalf("legacy settings remained flat = %#v", settings)
	}
	if settings["search"].(map[string]any)["sortBy"] != "SORT_BY_TITLE" {
		t.Fatalf("settings = %#v", settings)
	}
	if settings["download"].(map[string]any)["folderId"] != "42" {
		t.Fatalf("settings = %#v", settings)
	}
	if settings["catalog"].(map[string]any)["moviesSource"] != "MOVIES_SOURCE_YTS" {
		t.Fatalf("settings = %#v", settings)
	}

	wrapped, err := decodeUserSettingsRequest(app, `{"settings":{"filterNastyResults":true,"moviesSource":"MOVIES_SOURCE_YTS","downloadFolderId":42}}`)
	if err != nil {
		t.Fatalf("decodeUserSettingsRequest(wrapped) error = %v", err)
	}
	wrappedSettings := wrapped["settings"].(map[string]any)
	if wrappedSettings["search"].(map[string]any)["filterNastyResults"] != true {
		t.Fatalf("wrappedSettings = %#v", wrappedSettings)
	}

	if _, err := decodeUserSettingsRequest(app, `{"settings":{"filterNastyResults":true}}`); err == nil {
		t.Fatal("decodeUserSettingsRequest(partial) error = nil, want required domains error")
	}
}

func TestNormalizeIDsAndRendererMapping(t *testing.T) {
	t.Parallel()

	if _, err := normalizeFolderID("-1"); err == nil {
		t.Fatal("normalizeFolderID(-1) error = nil, want error")
	}
	if value, err := normalizeFolderID("42"); err != nil || value != 42 {
		t.Fatalf("normalizeFolderID(42) = %d, %v", value, err)
	}
	if _, err := normalizeTransferID("0"); err == nil {
		t.Fatal("normalizeTransferID(0) error = nil, want error")
	}
	if value, err := normalizeTransferID("42"); err != nil || value != 42 {
		t.Fatalf("normalizeTransferID(42) = %d, %v", value, err)
	}

	if prettyRendererForProcedure(procedureUserGetIndexers) == nil {
		t.Fatal("prettyRendererForProcedure(indexers) = nil")
	}
	if prettyRendererForProcedure(procedureUserGetMovies) == nil {
		t.Fatal("prettyRendererForProcedure(movies) = nil")
	}
	if prettyRendererForProcedure(procedureUserGetTVShows) == nil {
		t.Fatal("prettyRendererForProcedure(tv shows) = nil")
	}
	if prettyRendererForProcedure(procedureUserGetTVShowDetail) == nil {
		t.Fatal("prettyRendererForProcedure(tv show detail) = nil")
	}
	if prettyRendererForProcedure(procedureUserGetTVShowSeason) == nil {
		t.Fatal("prettyRendererForProcedure(tv show season) = nil")
	}
	if prettyRendererForProcedure(procedureUserGetTVShowEpisodeDownload) == nil {
		t.Fatal("prettyRendererForProcedure(tv show episode download) = nil")
	}
	if prettyRendererForProcedure(procedureUserGetTVShowSeasonDownloads) == nil {
		t.Fatal("prettyRendererForProcedure(tv show season downloads) = nil")
	}
	if prettyRendererForProcedure("unknown") != nil {
		t.Fatal("prettyRendererForProcedure(unknown) != nil")
	}
}

func TestNormalizeIMDbID(t *testing.T) {
	t.Parallel()

	if got, err := normalizeIMDbID(" tt0944947 "); err != nil || got != "tt0944947" {
		t.Fatalf("normalizeIMDbID() = %q, %v", got, err)
	}

	for _, raw := range []string{"", "bad", "ttabc", "tt123", "tt0944947?x=1", "tt0944947%2f"} {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			if _, err := normalizeIMDbID(raw); err == nil {
				t.Fatalf("normalizeIMDbID(%q) error = nil, want error", raw)
			}
		})
	}
}

func TestNormalizeEpisodeOrdinal(t *testing.T) {
	t.Parallel()

	if got, err := normalizeEpisodeOrdinal("1", "season"); err != nil || got != 1 {
		t.Fatalf("normalizeEpisodeOrdinal() = %d, %v", got, err)
	}
	if _, err := normalizeEpisodeOrdinal("0", "season"); err == nil {
		t.Fatal("normalizeEpisodeOrdinal(0) error = nil, want error")
	}
	if _, err := normalizeEpisodeOrdinal("abc", "episode"); err == nil {
		t.Fatal("normalizeEpisodeOrdinal(abc) error = nil, want error")
	}
}

func TestSettingsValidationHelpers(t *testing.T) {
	t.Parallel()

	if _, err := normalizeAPIBaseURL("https://api.chill.institute/path"); err == nil {
		t.Fatal("normalizeAPIBaseURL(path) error = nil, want error")
	}
	if _, err := normalizeAPIBaseURL("https://user@api.chill.institute"); err == nil {
		t.Fatal("normalizeAPIBaseURL(userinfo) error = nil, want error")
	}
	if got, err := normalizeAPIBaseURL("https://api.chill.institute/"); err != nil || got != "https://api.chill.institute" {
		t.Fatalf("normalizeAPIBaseURL() = %q, %v", got, err)
	}

	if _, err := normalizeSettingsKey("api_base_url"); err != nil {
		t.Fatalf("normalizeSettingsKey(api_base_url) error = %v", err)
	}
	if _, err := normalizeSettingsKey("missing"); err == nil {
		t.Fatal("normalizeSettingsKey(missing) error = nil, want error")
	}
}

func TestNormalizeIndexerID(t *testing.T) {
	t.Parallel()

	if got, err := normalizeIndexerID(" yts "); err != nil || got != "yts" {
		t.Fatalf("normalizeIndexerID() = %q, %v", got, err)
	}

	for _, raw := range []string{"", "bad\x00id", "../yts", "bad/id", "bad?id", "bad%2Fyts"} {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			if _, err := normalizeIndexerID(raw); err == nil {
				t.Fatalf("normalizeIndexerID(%q) error = nil, want error", raw)
			}
		})
	}
}

func TestNormalizeTransferURL(t *testing.T) {
	t.Parallel()

	if got, err := normalizeTransferURL(" magnet:?xt=urn:btih:test "); err != nil || got != "magnet:?xt=urn:btih:test" {
		t.Fatalf("normalizeTransferURL() = %q, %v", got, err)
	}

	if _, err := normalizeTransferURL(""); err == nil {
		t.Fatal("normalizeTransferURL(empty) error = nil, want error")
	}
	if _, err := normalizeTransferURL("bad\x00url"); err == nil {
		t.Fatal("normalizeTransferURL(control) error = nil, want error")
	}
}
