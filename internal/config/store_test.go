package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPathUsesXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/chilly-xdg")

	path, err := DefaultPath(defaultProfile)
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}

	want := filepath.Join("/tmp/chilly-xdg", appDirName, configFileName)
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestDefaultPathFallsBackToUserConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	originalUserConfigDir := userConfigDir
	userConfigDir = func() (string, error) { return "/tmp/chilly-user-config", nil }
	t.Cleanup(func() { userConfigDir = originalUserConfigDir })

	path, err := DefaultPath(defaultProfile)
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}

	want := filepath.Join("/tmp/chilly-user-config", appDirName, configFileName)
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

func TestNewStoreUsesDefaultPathWhenEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/chilly-store-default")

	store, err := NewStore(" ")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	want := filepath.Join("/tmp/chilly-store-default", appDirName, configFileName)
	if store.Path() != want {
		t.Fatalf("Path() = %q, want %q", store.Path(), want)
	}
}

func TestDefaultPathUsesProfilesSubdirectoryForNamedProfiles(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/chilly-xdg")

	path, err := DefaultPath("dev")
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}

	want := filepath.Join("/tmp/chilly-xdg", appDirName, profilesDirName, "dev", configFileName)
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestResolveProfileUsesDevDefaultForDevBuilds(t *testing.T) {
	t.Setenv(envProfile, "")

	profile, err := ResolveProfile("", true)
	if err != nil {
		t.Fatalf("ResolveProfile() error = %v", err)
	}
	if profile != devProfile {
		t.Fatalf("profile = %q, want %q", profile, devProfile)
	}
}

func TestResolveProfileUsesEnvironmentOverride(t *testing.T) {
	t.Setenv(envProfile, "production")

	profile, err := ResolveProfile("", false)
	if err != nil {
		t.Fatalf("ResolveProfile() error = %v", err)
	}
	if profile != "production" {
		t.Fatalf("profile = %q, want %q", profile, "production")
	}
}

func TestNormalizeProfileRejectsUnsafeValues(t *testing.T) {
	if _, err := NormalizeProfile("../prod"); err == nil {
		t.Fatal("expected invalid profile error")
	}
}

func TestDefaultProfileReturnsDefaultProfileName(t *testing.T) {
	if DefaultProfile() != defaultProfile {
		t.Fatalf("DefaultProfile() = %q, want %q", DefaultProfile(), defaultProfile)
	}
}
