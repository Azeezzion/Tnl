package update

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cli/cli/v2/pkg/cmd/extension"
	"github.com/cli/cli/v2/pkg/extensions"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCheckForUpdate(t *testing.T) {
	scenarios := []struct {
		Name           string
		CurrentVersion string
		LatestVersion  string
		LatestURL      string
		ExpectsResult  bool
	}{
		{
			Name:           "latest is newer",
			CurrentVersion: "v0.0.1",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "current is prerelease",
			CurrentVersion: "v1.0.0-pre.1",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "current is built from source",
			CurrentVersion: "v1.2.3-123-gdeadbeef",
			LatestVersion:  "v1.2.3",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  false,
		},
		{
			Name:           "current is built from source after a prerelease",
			CurrentVersion: "v1.2.3-rc.1-123-gdeadbeef",
			LatestVersion:  "v1.2.3",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "latest is newer than version build from source",
			CurrentVersion: "v1.2.3-123-gdeadbeef",
			LatestVersion:  "v1.2.4",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  true,
		},
		{
			Name:           "latest is current",
			CurrentVersion: "v1.0.0",
			LatestVersion:  "v1.0.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  false,
		},
		{
			Name:           "latest is older",
			CurrentVersion: "v0.10.0-pre.1",
			LatestVersion:  "v0.9.0",
			LatestURL:      "https://www.spacejam.com/archive/spacejam/movie/jam.htm",
			ExpectsResult:  false,
		},
	}

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			httpClient := &http.Client{}
			httpmock.ReplaceTripper(httpClient, reg)

			reg.Register(
				httpmock.REST("GET", "repos/OWNER/REPO/releases/latest"),
				httpmock.StringResponse(fmt.Sprintf(`{
					"tag_name": "%s",
					"html_url": "%s"
				}`, s.LatestVersion, s.LatestURL)),
			)

			rel, err := CheckForUpdate(context.TODO(), httpClient, tempFilePath(), "OWNER/REPO", s.CurrentVersion)
			if err != nil {
				t.Fatal(err)
			}

			if len(reg.Requests) != 1 {
				t.Fatalf("expected 1 HTTP request, got %d", len(reg.Requests))
			}
			requestPath := reg.Requests[0].URL.Path
			if requestPath != "/repos/OWNER/REPO/releases/latest" {
				t.Errorf("HTTP path: %q", requestPath)
			}

			if !s.ExpectsResult {
				if rel != nil {
					t.Fatal("expected no new release")
				}
				return
			}
			if rel == nil {
				t.Fatal("expected to report new release")
			}

			if rel.Version != s.LatestVersion {
				t.Errorf("Version: %q", rel.Version)
			}
			if rel.URL != s.LatestURL {
				t.Errorf("URL: %q", rel.URL)
			}
		})
	}
}

func TestCheckForExtensionUpdate(t *testing.T) {
	now := time.Date(2024, 12, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		extCurrentVersion   string
		extLatestVersion    string
		extIsLocal          bool
		extURL              string
		stateEntry          *StateEntry
		ri                  ReleaseInfo
		wantErr             bool
		expectedReleaseInfo *ReleaseInfo
		expectedStateEntry  *StateEntry
	}{
		{
			name:              "return latest release given extension is out of date and no state entry",
			extCurrentVersion: "v0.1.0",
			extLatestVersion:  "v1.0.0",
			extIsLocal:        false,
			extURL:            "http://example.com",
			stateEntry:        nil,
			expectedReleaseInfo: &ReleaseInfo{
				Version: "v1.0.0",
				URL:     "http://example.com",
			},
			expectedStateEntry: &StateEntry{
				CheckedForUpdateAt: now,
				LatestRelease: ReleaseInfo{
					Version: "v1.0.0",
					URL:     "http://example.com",
				},
			},
		},
		{
			name:              "return latest release given extension is out of date and state entry is old enough",
			extCurrentVersion: "v0.1.0",
			extLatestVersion:  "v1.0.0",
			extIsLocal:        false,
			extURL:            "http://example.com",
			stateEntry: &StateEntry{
				CheckedForUpdateAt: now.Add(-24 * time.Hour),
				LatestRelease: ReleaseInfo{
					Version: "v0.1.0",
					URL:     "http://example.com",
				},
			},
			expectedReleaseInfo: &ReleaseInfo{
				Version: "v1.0.0",
				URL:     "http://example.com",
			},
			expectedStateEntry: &StateEntry{
				CheckedForUpdateAt: now,
				LatestRelease: ReleaseInfo{
					Version: "v1.0.0",
					URL:     "http://example.com",
				},
			},
		},
		{
			name:              "return nothing given extension is out of date but state entry is too recent",
			extCurrentVersion: "v0.1.0",
			extLatestVersion:  "v1.0.0",
			extIsLocal:        false,
			extURL:            "http://example.com",
			stateEntry: &StateEntry{
				CheckedForUpdateAt: now.Add(-23 * time.Hour),
				LatestRelease: ReleaseInfo{
					Version: "v0.1.0",
					URL:     "http://example.com",
				},
			},
			expectedReleaseInfo: nil,
			expectedStateEntry: &StateEntry{
				CheckedForUpdateAt: now.Add(-23 * time.Hour),
				LatestRelease: ReleaseInfo{
					Version: "v0.1.0",
					URL:     "http://example.com",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := &extensions.ExtensionManagerMock{}

			ext := &extensions.ExtensionMock{
				CurrentVersionFunc: func() string {
					return tt.extCurrentVersion
				},
				LatestVersionFunc: func() string {
					return tt.extLatestVersion
				},
				IsLocalFunc: func() bool {
					return tt.extIsLocal
				},
				URLFunc: func() string {
					return tt.extURL
				},
			}

			// Ensure test is testing actual update available logic
			ext.UpdateAvailableFunc = func() bool {
				// Should this function be removed from the extension interface?
				return extension.UpdateAvailable(ext)
			}

			// Create state file for test as necessary
			stateFilePath := filepath.Join(t.TempDir(), "state.yml")
			if tt.stateEntry != nil {
				stateEntryYaml, err := yaml.Marshal(tt.stateEntry)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(stateFilePath, stateEntryYaml, 0644))
			}

			actual, err := CheckForExtensionUpdate(em, ext, stateFilePath, now)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Equal(t, tt.expectedReleaseInfo, actual)

				stateEntry, err := getStateEntry(stateFilePath)
				require.NoError(t, err)
				require.Equal(t, tt.expectedStateEntry, stateEntry)
			}
		})
	}
}

func tempFilePath() string {
	file, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(file.Name())
	return file.Name()
}
