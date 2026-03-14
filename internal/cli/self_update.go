package cli

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/chill-institute/cli/internal/buildinfo"
	"github.com/chill-institute/cli/internal/update"
	"github.com/spf13/cobra"
)

type releaseService interface {
	Latest(context.Context) (update.Release, error)
	ByTag(context.Context, string) (update.Release, error)
	Download(context.Context, string) ([]byte, error)
}

var (
	currentBuildInfo     = buildinfo.Current
	newReleaseService    = func() releaseService { return update.NewClient(http.DefaultClient) }
	currentExecutable    = os.Executable
	currentRuntimeGOOS   = runtime.GOOS
	currentRuntimeGOARCH = runtime.GOARCH
)

func newSelfUpdateCommand(app *appContext) *cobra.Command {
	var targetVersion string
	var checkOnly bool

	command := &cobra.Command{
		Use:   "self-update",
		Short: "Download and install the latest released CLI binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			service := newReleaseService()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			release, err := resolveRelease(ctx, service, targetVersion)
			if err != nil {
				return wrapInternalError("resolve_release_failed", "resolve release metadata", err)
			}

			current := currentBuildInfo()
			payload := map[string]any{
				"current_version": current.Version,
				"latest_version":  release.TagName,
				"checked":         true,
			}

			if checkOnly {
				payload["up_to_date"] = update.SameVersion(current.Version, release.TagName)
				return app.writeJSONPayload(payload)
			}

			if update.SameVersion(current.Version, release.TagName) {
				payload["updated"] = false
				payload["up_to_date"] = true
				return app.writeJSONPayload(payload)
			}

			if currentRuntimeGOOS == "windows" {
				return usageError("self_update_unsupported", "self-update is not supported on windows yet")
			}

			executablePath, err := currentExecutable()
			if err != nil {
				return wrapInternalError("resolve_executable_path_failed", "resolve current executable path", err)
			}

			asset, err := update.FindAsset(release, currentRuntimeGOOS, currentRuntimeGOARCH)
			if err != nil {
				return wrapInternalError("resolve_release_asset_failed", "resolve release asset", err)
			}
			archive, err := service.Download(ctx, asset.BrowserDownloadURL)
			if err != nil {
				return wrapInternalError("download_release_asset_failed", "download release asset", err)
			}
			binary, err := update.ExtractBinary(archive, currentRuntimeGOOS)
			if err != nil {
				return wrapInternalError("extract_release_asset_failed", "extract release asset", err)
			}

			mode := os.FileMode(0o755)
			if fileInfo, statErr := os.Stat(executablePath); statErr == nil {
				mode = fileInfo.Mode().Perm()
			}
			if err := update.ReplaceExecutable(executablePath, binary, mode); err != nil {
				return wrapInternalError("replace_executable_failed", "replace current executable", err)
			}

			payload["updated"] = true
			payload["up_to_date"] = false
			payload["path"] = executablePath
			payload["installed_version"] = release.TagName
			payload["asset"] = asset.Name
			return app.writeJSONPayload(payload)
		},
	}

	command.Flags().BoolVar(&checkOnly, "check", false, "check for a newer release without installing it")
	command.Flags().StringVar(&targetVersion, "version", "", "specific release tag to install, for example v0.1.0")
	return command
}

func resolveRelease(ctx context.Context, service releaseService, version string) (update.Release, error) {
	if update.NormalizeVersion(version) == "" {
		return service.Latest(ctx)
	}
	return service.ByTag(ctx, version)
}
