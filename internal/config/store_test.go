package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPathUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/chilly-xdg")

	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}

	want := filepath.Join("/tmp/chilly-xdg", appDirName, configFileName)
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestLoadMissingConfigReturnsDefaults(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIBaseURL != defaultAPIBase {
		t.Fatalf("APIBaseURL = %q, want %q", cfg.APIBaseURL, defaultAPIBase)
	}
	if cfg.AuthToken != "" {
		t.Fatalf("AuthToken = %q, want empty", cfg.AuthToken)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "cfg.json")
	store, err := NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	saved := Config{
		APIBaseURL: " https://chill.example ",
		AuthToken:  " token-123 ",
	}
	if err := store.Save(saved); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.APIBaseURL != "https://chill.example" {
		t.Fatalf("APIBaseURL = %q, want %q", loaded.APIBaseURL, "https://chill.example")
	}
	if loaded.AuthToken != "token-123" {
		t.Fatalf("AuthToken = %q, want %q", loaded.AuthToken, "token-123")
	}

	stat, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if stat.Mode().Perm() != configFilePerm {
		t.Fatalf("perm = %o, want %o", stat.Mode().Perm(), configFilePerm)
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(raw), "\n") {
		t.Fatal("expected saved config to be pretty-printed")
	}
}
