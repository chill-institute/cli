package buildinfo

import "testing"

func TestCurrentNormalizesEmptyValues(t *testing.T) {
	originalVersion := version
	originalCommit := commit
	originalDate := date
	version = " "
	commit = ""
	date = "\t"
	t.Cleanup(func() {
		version = originalVersion
		commit = originalCommit
		date = originalDate
	})

	info := Current()
	if info.Version != "dev" {
		t.Fatalf("Version = %q", info.Version)
	}
	if info.Commit != "unknown" {
		t.Fatalf("Commit = %q", info.Commit)
	}
	if info.BuildDate != "unknown" {
		t.Fatalf("BuildDate = %q", info.BuildDate)
	}
}

func TestInfoIsDev(t *testing.T) {
	t.Parallel()

	if !(Info{Version: "dev"}).IsDev() {
		t.Fatal("expected dev build")
	}
	if (Info{Version: "v1.2.3"}).IsDev() {
		t.Fatal("release build incorrectly marked dev")
	}
}
