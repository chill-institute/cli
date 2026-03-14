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

func TestParseFieldSelectionRejectsInvalidPath(t *testing.T) {
	t.Parallel()

	if _, err := parseFieldSelection("results..title"); err == nil {
		t.Fatal("expected error")
	}
}
