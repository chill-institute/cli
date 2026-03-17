package cli

import (
	"testing"
)

func TestNormalizeUserSettingsPatch(t *testing.T) {
	t.Parallel()

	patch, err := normalizeUserSettingsPatch("show-top-movies", "true")
	if err != nil {
		t.Fatalf("normalizeUserSettingsPatch() error = %v", err)
	}
	if patch.Field != "showTopMovies" || patch.Value != true {
		t.Fatalf("patch = %#v", patch)
	}

	enumPatch, err := normalizeUserSettingsPatch("sort-by", "uploaded-at")
	if err != nil {
		t.Fatalf("normalizeUserSettingsPatch(enum) error = %v", err)
	}
	if enumPatch.Value != "SORT_BY_UPLOADED_AT" {
		t.Fatalf("enumPatch = %#v", enumPatch)
	}

	if _, err := normalizeUserSettingsPatch("missing-field", "x"); err == nil {
		t.Fatal("normalizeUserSettingsPatch() error = nil, want unsupported field")
	}
}

func TestNormalizePatchFieldNameAndKebabCase(t *testing.T) {
	t.Parallel()

	if got := normalizePatchFieldName("show-top-movies"); got != "showTopMovies" {
		t.Fatalf("normalizePatchFieldName() = %q", got)
	}
	if got := normalizePatchFieldName("showTopMovies"); got != "showTopMovies" {
		t.Fatalf("normalizePatchFieldName(camel) = %q", got)
	}
	if got := kebabCase("showTopMovies"); got != "show-top-movies" {
		t.Fatalf("kebabCase() = %q", got)
	}
}

func TestNormalizeNullableInt64Value(t *testing.T) {
	t.Parallel()

	if value, err := normalizeNullableInt64Value("42"); err != nil || value != "42" {
		t.Fatalf("normalizeNullableInt64Value(42) = %#v, %v", value, err)
	}
	if value, err := normalizeNullableInt64Value("null"); err != nil || value != nil {
		t.Fatalf("normalizeNullableInt64Value(null) = %#v, %v", value, err)
	}
	if value, err := normalizeNullableInt64Value("none"); err != nil || value != nil {
		t.Fatalf("normalizeNullableInt64Value(none) = %#v, %v", value, err)
	}
	if _, err := normalizeNullableInt64Value(""); err == nil {
		t.Fatal("normalizeNullableInt64Value(empty) error = nil, want error")
	}
	if _, err := normalizeNullableInt64Value("nope"); err == nil {
		t.Fatal("normalizeNullableInt64Value(nope) error = nil, want error")
	}
}

func TestNormalizeEnumValue(t *testing.T) {
	t.Parallel()

	normalize := normalizeEnumValue(map[string]string{"asc": "ASC"})
	if value, err := normalize("ASC"); err != nil || value != "ASC" {
		t.Fatalf("normalizeEnumValue() = %#v, %v", value, err)
	}
	if _, err := normalize("desc"); err == nil {
		t.Fatal("normalizeEnumValue(desc) error = nil, want error")
	}
}

func TestApplyUserSettingsPatchAndCloneJSONObject(t *testing.T) {
	t.Parallel()

	source := map[string]any{
		"settings": map[string]any{
			"nested": "value",
		},
		"items": []any{"a", "b"},
	}

	cloned := cloneJSONObject(source)
	clonedSettings := cloned["settings"].(map[string]any)
	clonedSettings["nested"] = "changed"
	if source["settings"].(map[string]any)["nested"] != "value" {
		t.Fatalf("source mutated = %#v", source)
	}

	patched := applyUserSettingsPatch(source, userSettingsPatch{
		Field: "showTopMovies",
		Value: true,
	})
	if patched["showTopMovies"] != true {
		t.Fatalf("patched = %#v", patched)
	}
}
