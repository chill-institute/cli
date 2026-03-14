package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	repoOwner      = "chill-institute"
	repoName       = "cli"
	binaryName     = "chilly"
	defaultAPIBase = "https://api.github.com"
)

type Release struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Assets  []ReleaseAsset `json:"assets"`
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		baseURL:    defaultAPIBase,
		httpClient: client,
	}
}

func (client Client) Latest(ctx context.Context) (Release, error) {
	return client.fetchRelease(ctx, "/repos/"+repoOwner+"/"+repoName+"/releases/latest")
}

func (client Client) ByTag(ctx context.Context, version string) (Release, error) {
	normalized := NormalizeVersion(version)
	if normalized == "" {
		return Release{}, errors.New("release version is required")
	}
	return client.fetchRelease(ctx, "/repos/"+repoOwner+"/"+repoName+"/releases/tags/"+normalized)
}

func (client Client) Download(ctx context.Context, downloadURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(downloadURL), nil)
	if err != nil {
		return nil, fmt.Errorf("build download request: %w", err)
	}
	request.Header.Set("Accept", "application/octet-stream")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("download release asset: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return nil, fmt.Errorf("download release asset: unexpected status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read release asset: %w", err)
	}
	return payload, nil
}

func NormalizeVersion(version string) string {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "v") {
		return trimmed
	}
	return "v" + trimmed
}

func SameVersion(left string, right string) bool {
	normalizedLeft := NormalizeVersion(left)
	normalizedRight := NormalizeVersion(right)
	return normalizedLeft != "" && normalizedLeft == normalizedRight
}

func ArchiveName(version string, goos string, goarch string) (string, error) {
	normalizedVersion := NormalizeVersion(version)
	if normalizedVersion == "" {
		return "", errors.New("release version is required")
	}

	trimmedOS := strings.TrimSpace(goos)
	trimmedArch := strings.TrimSpace(goarch)
	if trimmedOS == "" || trimmedArch == "" {
		return "", errors.New("target os and arch are required")
	}

	switch trimmedOS {
	case "darwin", "linux":
		return fmt.Sprintf("%s_%s_%s_%s.tar.gz", binaryName, normalizedVersion, trimmedOS, trimmedArch), nil
	case "windows":
		return fmt.Sprintf("%s_%s_%s_%s.zip", binaryName, normalizedVersion, trimmedOS, trimmedArch), nil
	default:
		return "", fmt.Errorf("unsupported target os %q", goos)
	}
}

func FindAsset(release Release, goos string, goarch string) (ReleaseAsset, error) {
	expectedName, err := ArchiveName(release.TagName, goos, goarch)
	if err != nil {
		return ReleaseAsset{}, err
	}

	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset, nil
		}
	}

	return ReleaseAsset{}, fmt.Errorf("release asset %q not found", expectedName)
}

func ExtractBinary(archive []byte, goos string) ([]byte, error) {
	switch strings.TrimSpace(goos) {
	case "darwin", "linux":
		return extractTarGZBinary(archive)
	case "windows":
		return extractZipBinary(archive)
	default:
		return nil, fmt.Errorf("unsupported target os %q", goos)
	}
}

func ReplaceExecutable(path string, binary []byte, mode os.FileMode) error {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return errors.New("executable path is required")
	}
	if len(binary) == 0 {
		return errors.New("binary payload is required")
	}
	if mode == 0 {
		mode = 0o755
	}

	dir := filepath.Dir(trimmedPath)
	tmpFile, err := os.CreateTemp(dir, binaryName+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp executable: %w", err)
	}
	tmpPath := tmpFile.Name()
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	if _, err := tmpFile.Write(binary); err != nil {
		_ = tmpFile.Close()
		cleanup()
		return fmt.Errorf("write temp executable: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp executable: %w", err)
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp executable: %w", err)
	}
	if err := os.Rename(tmpPath, trimmedPath); err != nil {
		cleanup()
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}

func (client Client) fetchRelease(ctx context.Context, path string) (Release, error) {
	endpoint, err := url.JoinPath(client.baseURL, path)
	if err != nil {
		return Release{}, fmt.Errorf("build release endpoint: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Release{}, fmt.Errorf("build release request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return Release{}, fmt.Errorf("fetch release metadata: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return Release{}, fmt.Errorf("read release metadata: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Release{}, fmt.Errorf("fetch release metadata: unexpected status %d: %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}

	var release Release
	if err := json.Unmarshal(payload, &release); err != nil {
		return Release{}, fmt.Errorf("decode release metadata: %w", err)
	}
	release.TagName = NormalizeVersion(release.TagName)
	return release, nil
}

func extractTarGZBinary(archive []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer func() {
		_ = gzipReader.Close()
	}()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar.gz archive: %w", err)
		}
		if header.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(header.Name) != binaryName {
			continue
		}
		payload, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, fmt.Errorf("read binary payload: %w", err)
		}
		return payload, nil
	}

	return nil, errors.New("binary not found in tar.gz archive")
}

func extractZipBinary(archive []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, fmt.Errorf("open zip archive: %w", err)
	}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(file.Name) != binaryName+".exe" {
			continue
		}
		handle, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open binary payload: %w", err)
		}
		payload, readErr := io.ReadAll(handle)
		closeErr := handle.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read binary payload: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close binary payload: %w", closeErr)
		}
		return payload, nil
	}

	return nil, errors.New("binary not found in zip archive")
}
