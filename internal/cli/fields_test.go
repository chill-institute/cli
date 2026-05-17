package cli

import (
	"reflect"
	"testing"
)

func TestParseFieldSelectionAppliesNestedPaths(t *testing.T) {
	t.Parallel()

	selection, err := parseFieldSelection("results.title, results.magnetLink, request_id")
	if err != nil {
		t.Fatalf("parseFieldSelection() error = %v", err)
	}

	input := map[string]any{
		"request_id": "req-123",
		"results": []any{
			map[string]any{
				"title":      "Dune",
				"magnetLink": "magnet:?xt=urn:btih:dune",
				"size":       "1.4 GB",
			},
		},
		"ignored": true,
	}

	want := map[string]any{
		"request_id": "req-123",
		"results": []any{
			map[string]any{
				"title":      "Dune",
				"magnetLink": "magnet:?xt=urn:btih:dune",
			},
		},
	}

	got := selection.apply(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selection.apply() = %#v, want %#v", got, want)
	}
}

func TestParseFieldSelectionAcceptsProtoSnakeCaseForJSONNames(t *testing.T) {
	t.Parallel()

	selection, err := parseFieldSelection("results.release_info.bit_depth,results.uploaded_at")
	if err != nil {
		t.Fatalf("parseFieldSelection() error = %v", err)
	}

	input := map[string]any{
		"results": []any{
			map[string]any{
				"title":      "Dune",
				"uploadedAt": "2026-05-17T10:00:00Z",
				"releaseInfo": map[string]any{
					"bitDepth": "10-bit",
					"edition":  "IMAX",
				},
			},
		},
	}

	want := map[string]any{
		"results": []any{
			map[string]any{
				"uploadedAt": "2026-05-17T10:00:00Z",
				"releaseInfo": map[string]any{
					"bitDepth": "10-bit",
				},
			},
		},
	}

	got := selection.apply(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selection.apply() = %#v, want %#v", got, want)
	}
}

func TestParseFieldSelectionAcceptsJSONNamesForSnakeCasePayloads(t *testing.T) {
	t.Parallel()

	selection, err := parseFieldSelection("releaseInfo.bitDepth")
	if err != nil {
		t.Fatalf("parseFieldSelection() error = %v", err)
	}

	input := map[string]any{
		"release_info": map[string]any{
			"bit_depth": "10-bit",
			"edition":   "IMAX",
		},
	}

	want := map[string]any{
		"release_info": map[string]any{
			"bit_depth": "10-bit",
		},
	}

	got := selection.apply(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selection.apply() = %#v, want %#v", got, want)
	}
}

func TestParseFieldSelectionRejectsInvalidPath(t *testing.T) {
	t.Parallel()

	if _, err := parseFieldSelection("results..title"); err == nil {
		t.Fatal("expected error")
	}
}
