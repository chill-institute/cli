package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/chill-institute/cli/internal/buildinfo"
	"github.com/chill-institute/cli/internal/update"
)

type stubReleaseService struct {
	latestRelease update.Release
	tagRelease    update.Release
}

func (service *stubReleaseService) Latest(context.Context) (update.Release, error) {
	return service.latestRelease, nil
}

func (service *stubReleaseService) ByTag(context.Context, string) (update.Release, error) {
	return service.tagRelease, nil
}

func (service *stubReleaseService) Download(context.Context, string) ([]byte, error) {
	return nil, errors.New("unexpected download")
}

func TestSelfUpdateCheckReportsReleaseWithoutInstalling(t *testing.T) {
	restoreBuildInfo := currentBuildInfo
	restoreReleaseService := newReleaseService
	restoreGOOS := currentRuntimeGOOS
	restoreGOARCH := currentRuntimeGOARCH
	currentBuildInfo = func() buildinfo.Info { return buildinfo.Info{Version: "v1.0.0"} }
	newReleaseService = func() releaseService {
		return &stubReleaseService{
			latestRelease: update.Release{TagName: "v1.2.0"},
		}
	}
	currentRuntimeGOOS = "darwin"
	currentRuntimeGOARCH = "arm64"
	t.Cleanup(func() {
		currentBuildInfo = restoreBuildInfo
		newReleaseService = restoreReleaseService
		currentRuntimeGOOS = restoreGOOS
		currentRuntimeGOARCH = restoreGOARCH
	})

	stdout := &bytes.Buffer{}
	command := newSelfUpdateCommand(&appContext{
		opts:   &appOptions{output: outputJSON},
		stdin:  strings.NewReader(""),
		stdout: stdout,
		stderr: &bytes.Buffer{},
	})
	command.SetArgs([]string{"--check"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if output["latest_version"] != "v1.2.0" {
		t.Fatalf("latest_version = %v", output["latest_version"])
	}
	if output["up_to_date"] != false {
		t.Fatalf("up_to_date = %v", output["up_to_date"])
	}
}
