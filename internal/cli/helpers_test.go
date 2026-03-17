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
	if setRequest["settings"].(map[string]any)["downloadFolderId"] != "42" {
		t.Fatalf("setRequest = %#v", setRequest)
	}

	clearRequest, err := resolveDownloadFolderClearRequest(app, `{"settings":{"downloadFolderId":null}}`)
	if err != nil {
		t.Fatalf("resolveDownloadFolderClearRequest() error = %v", err)
	}
	if value := clearRequest["settings"].(map[string]any)["downloadFolderId"]; value != nil {
		t.Fatalf("clearRequest = %#v", clearRequest)
	}

	if _, err := resolveDownloadFolderRequest(app, `{}`, false); err == nil {
		t.Fatal("resolveDownloadFolderRequest() error = nil, want missing field error")
	}
	if _, err := resolveDownloadFolderClearRequest(app, `{"settings":{"downloadFolderId":42}}`); err == nil {
		t.Fatal("resolveDownloadFolderClearRequest() error = nil, want null requirement error")
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
	if prettyRendererForProcedure("unknown") != nil {
		t.Fatal("prettyRendererForProcedure(unknown) != nil")
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
