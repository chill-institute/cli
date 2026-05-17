package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreatesRootAndPlatformPackages(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "dist")
	outDir := filepath.Join(dir, "npm")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}

	artifacts := make([]artifact, 0, len(targets))
	for _, target := range targets {
		binaryPath := filepath.Join(distDir, "build", target.suffix, target.binaryFile)
		if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(binaryPath, []byte("binary"), 0o755); err != nil {
			t.Fatal(err)
		}
		artifacts = append(artifacts, artifact{
			Name:   target.binaryFile,
			Path:   binaryPath,
			GoOS:   target.goOS,
			GoArch: target.goArch,
			Type:   "Binary",
		})
	}
	writeJSONFixture(t, filepath.Join(distDir, "artifacts.json"), artifacts)

	if err := run(options{distDir: distDir, outDir: outDir, version: "v1.2.3"}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	root := readPackageJSON(t, filepath.Join(outDir, "cli", "package.json"))
	if root.Name != rootPackageName {
		t.Fatalf("root package name = %q, want %q", root.Name, rootPackageName)
	}
	if root.Version != "1.2.3" {
		t.Fatalf("root package version = %q, want 1.2.3", root.Version)
	}
	if root.Bin[binaryName] != "bin/chilly.js" {
		t.Fatalf("root bin = %#v, want chilly launcher", root.Bin)
	}
	if len(root.OptionalDependencies) != len(targets) {
		t.Fatalf("optional dependency count = %d, want %d", len(root.OptionalDependencies), len(targets))
	}

	launcherPath := filepath.Join(outDir, "cli", "bin", "chilly.js")
	launcher, err := os.ReadFile(launcherPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(launcher), "@chill-institute/cli-darwin-arm64") {
		t.Fatalf("launcher does not contain platform package map:\n%s", launcher)
	}
	assertExecutable(t, launcherPath)

	platform := readPackageJSON(t, filepath.Join(outDir, "cli-darwin-arm64", "package.json"))
	if platform.Name != "@chill-institute/cli-darwin-arm64" {
		t.Fatalf("platform package name = %q", platform.Name)
	}
	if len(platform.OS) != 1 || platform.OS[0] != "darwin" {
		t.Fatalf("platform os = %#v, want darwin", platform.OS)
	}
	if len(platform.CPU) != 1 || platform.CPU[0] != "arm64" {
		t.Fatalf("platform cpu = %#v, want arm64", platform.CPU)
	}
	assertExecutable(t, filepath.Join(outDir, "cli-darwin-arm64", "bin", "chilly"))
}

func TestRunUsesMetadataVersionWhenVersionFlagIsEmpty(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "dist")
	outDir := filepath.Join(dir, "npm")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}

	artifacts := make([]artifact, 0, len(targets))
	for _, target := range targets {
		binaryPath := filepath.Join(distDir, target.suffix, target.binaryFile)
		if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(binaryPath, []byte("binary"), 0o755); err != nil {
			t.Fatal(err)
		}
		artifacts = append(artifacts, artifact{
			Path:   binaryPath,
			GoOS:   target.goOS,
			GoArch: target.goArch,
			Type:   "Binary",
		})
	}
	writeJSONFixture(t, filepath.Join(distDir, "artifacts.json"), artifacts)
	writeJSONFixture(t, filepath.Join(distDir, "metadata.json"), metadata{Version: "v2.3.4"})

	if err := run(options{distDir: distDir, outDir: outDir}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	root := readPackageJSON(t, filepath.Join(outDir, "cli", "package.json"))
	if root.Version != "2.3.4" {
		t.Fatalf("root package version = %q, want 2.3.4", root.Version)
	}
}

func TestRunRejectsMissingOptions(t *testing.T) {
	if err := run(options{outDir: t.TempDir(), version: "1.0.0"}); err == nil {
		t.Fatal("run() error = nil, want missing dist directory error")
	}
	if err := run(options{distDir: t.TempDir(), version: "1.0.0"}); err == nil {
		t.Fatal("run() error = nil, want missing output directory error")
	}
}

func TestRunRejectsMissingBinaryArtifact(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "dist")
	outDir := filepath.Join(dir, "npm")
	if err := os.MkdirAll(distDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSONFixture(t, filepath.Join(distDir, "artifacts.json"), []artifact{
		{Path: filepath.Join(distDir, "chilly"), GoOS: "darwin", GoArch: "arm64", Type: "Binary"},
	})

	err := run(options{distDir: distDir, outDir: outDir, version: "1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "missing GoReleaser binary artifact") {
		t.Fatalf("run() error = %v, want missing binary artifact error", err)
	}
}

func TestResolveVersionRejectsMissingMetadataVersion(t *testing.T) {
	dir := t.TempDir()
	writeJSONFixture(t, filepath.Join(dir, "metadata.json"), metadata{})

	_, err := resolveVersion(options{distDir: dir})
	if err == nil || !strings.Contains(err.Error(), "metadata version is empty") {
		t.Fatalf("resolveVersion() error = %v, want empty metadata version error", err)
	}
}

func TestReadArtifactsRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifacts.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := readArtifacts(path)
	if err == nil || !strings.Contains(err.Error(), "parse artifacts") {
		t.Fatalf("readArtifacts() error = %v, want parse error", err)
	}
}

func TestBinaryArtifactsFiltersNonBinaryEntries(t *testing.T) {
	binaries := binaryArtifacts([]artifact{
		{Path: "archive.tar.gz", GoOS: "darwin", GoArch: "arm64", Type: "Archive"},
		{Path: "chilly", GoOS: "darwin", GoArch: "arm64", Type: "Binary"},
		{Path: "", GoOS: "linux", GoArch: "arm64", Type: "Binary"},
	})

	if len(binaries) != 1 || binaries["darwin/arm64"] != "chilly" {
		t.Fatalf("binaryArtifacts() = %#v, want one darwin/arm64 binary", binaries)
	}
}

func TestResetOutputDirRejectsUnsafePath(t *testing.T) {
	if err := resetOutputDir("."); err == nil {
		t.Fatal("resetOutputDir() error = nil, want unsafe path error")
	}
}

func TestWriteJSONRejectsUnmarshalableValue(t *testing.T) {
	err := writeJSON(filepath.Join(t.TempDir(), "package.json"), make(chan struct{}))
	if err == nil || !strings.Contains(err.Error(), "marshal") {
		t.Fatalf("writeJSON() error = %v, want marshal error", err)
	}
}

func TestCopyFileRejectsMissingSource(t *testing.T) {
	err := copyFile(filepath.Join(t.TempDir(), "missing"), filepath.Join(t.TempDir(), "out"), 0o755)
	if err == nil {
		t.Fatal("copyFile() error = nil, want missing source error")
	}
}

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readPackageJSON(t *testing.T, path string) packageJSON {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatal(err)
	}
	return pkg
}

func assertExecutable(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("%s is not executable: %s", path, info.Mode())
	}
}
