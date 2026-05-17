package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	rootPackageName = "@chill-institute/cli"
	binaryName      = "chilly"
	repositoryURL   = "git+https://github.com/chill-institute/chill-cli.git"
	homepageURL     = "https://github.com/chill-institute/chill-cli#readme"
)

type options struct {
	distDir string
	outDir  string
	version string
}

type artifact struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	GoOS   string `json:"goos"`
	GoArch string `json:"goarch"`
	Type   string `json:"type"`
}

type metadata struct {
	Version string `json:"version"`
}

type target struct {
	goOS       string
	goArch     string
	npmOS      string
	npmArch    string
	suffix     string
	binaryFile string
}

type packageJSON struct {
	Name                 string            `json:"name"`
	Version              string            `json:"version"`
	Description          string            `json:"description"`
	License              string            `json:"license"`
	Repository           repository        `json:"repository"`
	Bugs                 bugs              `json:"bugs"`
	Homepage             string            `json:"homepage"`
	Keywords             []string          `json:"keywords,omitempty"`
	Bin                  map[string]string `json:"bin,omitempty"`
	Files                []string          `json:"files"`
	OS                   []string          `json:"os,omitempty"`
	CPU                  []string          `json:"cpu,omitempty"`
	OptionalDependencies map[string]string `json:"optionalDependencies,omitempty"`
	PublishConfig        publishConfig     `json:"publishConfig"`
}

type repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type bugs struct {
	URL string `json:"url"`
}

type publishConfig struct {
	Access string `json:"access"`
}

var targets = []target{
	{goOS: "darwin", goArch: "amd64", npmOS: "darwin", npmArch: "x64", suffix: "darwin-x64", binaryFile: binaryName},
	{goOS: "darwin", goArch: "arm64", npmOS: "darwin", npmArch: "arm64", suffix: "darwin-arm64", binaryFile: binaryName},
	{goOS: "linux", goArch: "amd64", npmOS: "linux", npmArch: "x64", suffix: "linux-x64", binaryFile: binaryName},
	{goOS: "linux", goArch: "arm64", npmOS: "linux", npmArch: "arm64", suffix: "linux-arm64", binaryFile: binaryName},
	{goOS: "windows", goArch: "amd64", npmOS: "win32", npmArch: "x64", suffix: "win32-x64", binaryFile: binaryName + ".exe"},
	{goOS: "windows", goArch: "arm64", npmOS: "win32", npmArch: "arm64", suffix: "win32-arm64", binaryFile: binaryName + ".exe"},
}

func main() {
	cfg := options{}
	flag.StringVar(&cfg.distDir, "dist", "dist", "GoReleaser dist directory")
	flag.StringVar(&cfg.outDir, "out", "dist/npm", "output directory for npm packages")
	flag.StringVar(&cfg.version, "version", "", "npm package version; defaults to dist/metadata.json")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "prepare npm packages: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg options) error {
	if cfg.distDir == "" {
		return errors.New("dist directory is required")
	}
	if cfg.outDir == "" {
		return errors.New("output directory is required")
	}

	version, err := resolveVersion(cfg)
	if err != nil {
		return err
	}

	artifacts, err := readArtifacts(filepath.Join(cfg.distDir, "artifacts.json"))
	if err != nil {
		return err
	}

	binaries := binaryArtifacts(artifacts)
	if err := resetOutputDir(cfg.outDir); err != nil {
		return err
	}

	if err := writeRootPackage(cfg.outDir, version); err != nil {
		return err
	}
	for _, t := range targets {
		source, ok := binaries[t.goOS+"/"+t.goArch]
		if !ok {
			return fmt.Errorf("missing GoReleaser binary artifact for %s/%s", t.goOS, t.goArch)
		}
		if err := writePlatformPackage(cfg.outDir, version, t, source); err != nil {
			return err
		}
	}

	return nil
}

func resolveVersion(cfg options) (string, error) {
	version := strings.TrimPrefix(strings.TrimSpace(cfg.version), "v")
	if version != "" {
		return version, nil
	}

	data, err := os.ReadFile(filepath.Join(cfg.distDir, "metadata.json"))
	if err != nil {
		return "", fmt.Errorf("read metadata version: %w", err)
	}
	var meta metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return "", fmt.Errorf("parse metadata version: %w", err)
	}
	version = strings.TrimPrefix(strings.TrimSpace(meta.Version), "v")
	if version == "" {
		return "", errors.New("metadata version is empty")
	}
	return version, nil
}

func readArtifacts(path string) ([]artifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read artifacts: %w", err)
	}
	var artifacts []artifact
	if err := json.Unmarshal(data, &artifacts); err != nil {
		return nil, fmt.Errorf("parse artifacts: %w", err)
	}
	return artifacts, nil
}

func binaryArtifacts(artifacts []artifact) map[string]string {
	binaries := map[string]string{}
	for _, item := range artifacts {
		if item.Type != "Binary" || item.GoOS == "" || item.GoArch == "" || item.Path == "" {
			continue
		}
		binaries[item.GoOS+"/"+item.GoArch] = item.Path
	}
	return binaries
}

func resetOutputDir(path string) error {
	clean := filepath.Clean(path)
	if clean == "." || clean == string(filepath.Separator) {
		return fmt.Errorf("refusing to clean unsafe output directory %q", path)
	}
	if err := os.RemoveAll(clean); err != nil {
		return fmt.Errorf("clean output directory: %w", err)
	}
	if err := os.MkdirAll(clean, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	return nil
}

func writeRootPackage(outDir string, version string) error {
	packageDir := filepath.Join(outDir, "cli")
	if err := os.MkdirAll(filepath.Join(packageDir, "bin"), 0o755); err != nil {
		return fmt.Errorf("create root package: %w", err)
	}

	optionalDependencies := map[string]string{}
	for _, t := range targets {
		optionalDependencies[platformPackageName(t)] = version
	}

	pkg := basePackage(rootPackageName, version, "Agent-first command-line client for chill.institute")
	pkg.Bin = map[string]string{binaryName: "bin/chilly.js"}
	pkg.Files = []string{"bin"}
	pkg.Keywords = []string{"chill.institute", "cli", "chilly"}
	pkg.OptionalDependencies = optionalDependencies

	if err := writeJSON(filepath.Join(packageDir, "package.json"), pkg); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(packageDir, "README.md"), []byte(rootReadme()), 0o644); err != nil {
		return fmt.Errorf("write root README: %w", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "bin", "chilly.js"), []byte(wrapperScript()), 0o755); err != nil {
		return fmt.Errorf("write launcher: %w", err)
	}
	return nil
}

func writePlatformPackage(outDir string, version string, t target, source string) error {
	packageDir := filepath.Join(outDir, platformDirName(t))
	binDir := filepath.Join(packageDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("create platform package %s: %w", t.suffix, err)
	}

	pkg := basePackage(platformPackageName(t), version, fmt.Sprintf("chilly binary for %s %s", t.npmOS, t.npmArch))
	pkg.Files = []string{"bin"}
	pkg.OS = []string{t.npmOS}
	pkg.CPU = []string{t.npmArch}

	if err := writeJSON(filepath.Join(packageDir, "package.json"), pkg); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(packageDir, "README.md"), []byte(platformReadme(t)), 0o644); err != nil {
		return fmt.Errorf("write platform README: %w", err)
	}
	if err := copyFile(source, filepath.Join(binDir, t.binaryFile), 0o755); err != nil {
		return fmt.Errorf("copy platform binary %s: %w", t.suffix, err)
	}
	return nil
}

func basePackage(name string, version string, description string) packageJSON {
	return packageJSON{
		Name:        name,
		Version:     version,
		Description: description,
		License:     "MIT",
		Repository: repository{
			Type: "git",
			URL:  repositoryURL,
		},
		Bugs: bugs{
			URL: "https://github.com/chill-institute/chill-cli/issues",
		},
		Homepage: homepageURL,
		PublishConfig: publishConfig{
			Access: "public",
		},
	}
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func copyFile(source string, destination string, mode os.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func platformPackageName(t target) string {
	return rootPackageName + "-" + t.suffix
}

func platformDirName(t target) string {
	return "cli-" + t.suffix
}

func rootReadme() string {
	return `# @chill-institute/cli

npm distribution for the ` + "`chilly`" + ` command-line client.

` + "```sh" + `
npm install -g @chill-institute/cli
chilly version
` + "```" + `
`
}

func platformReadme(t target) string {
	return "# " + platformPackageName(t) + "\n\n" +
		"Platform binary package for `@chill-institute/cli` on `" + t.npmOS + "-" + t.npmArch + "`.\n"
}

func wrapperScript() string {
	entries := make([]string, 0, len(targets))
	for _, t := range targets {
		entries = append(entries, fmt.Sprintf(
			"  %q: { packageName: %q, binaryName: %q }",
			t.npmOS+" "+t.npmArch,
			platformPackageName(t),
			t.binaryFile,
		))
	}
	sort.Strings(entries)

	return `#!/usr/bin/env node
"use strict";

const { spawn } = require("node:child_process");
const { dirname, join } = require("node:path");

const targets = {
` + strings.Join(entries, ",\n") + `
};

const targetKey = process.platform + " " + process.arch;
const target = targets[targetKey];
if (!target) {
  console.error("chilly is not available for " + process.platform + "-" + process.arch);
  process.exit(1);
}

let packageRoot;
try {
  packageRoot = dirname(require.resolve(target.packageName + "/package.json"));
} catch (error) {
  console.error("Missing optional dependency " + target.packageName + "; reinstall @chill-institute/cli with optional dependencies enabled.");
  process.exit(1);
}

const child = spawn(join(packageRoot, "bin", target.binaryName), process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: false
});

child.on("error", (error) => {
  console.error("Failed to start chilly: " + error.message);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code ?? 1);
});
`
}
