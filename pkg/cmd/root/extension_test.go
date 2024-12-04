package root_test

import (
	"io"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/update"
	"github.com/cli/cli/v2/pkg/cmd/root"
	"github.com/cli/cli/v2/pkg/extensions"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdExtension_Updates(t *testing.T) {
	tests := []struct {
		name               string
		extCurrentVersion  string
		extFullName        string
		extIsPinned        bool
		extLatestVersion   string
		extName            string
		extUpdateAvailable bool
		extURL             string
		wantStderr         string
	}{
		{
			name:               "no update available",
			extName:            "no-update",
			extFullName:        "gh-no-update",
			extUpdateAvailable: false,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "1.0.0",
			extURL:             "https//github.com/dne/no-update",
		},
		{
			name:               "major update",
			extName:            "major-update",
			extFullName:        "gh-major-update",
			extUpdateAvailable: true,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "2.0.0",
			extURL:             "https//github.com/dne/major-update",
			wantStderr: heredoc.Doc(`
				A new release of major-update is available: 1.0.0 → 2.0.0
				To upgrade, run: gh extension upgrade major-update
				https//github.com/dne/major-update
			`),
		},
		{
			name:               "major update, pinned",
			extName:            "major-update-pin",
			extFullName:        "gh-major-update-pin",
			extUpdateAvailable: true,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "2.0.0",
			extIsPinned:        true,
			extURL:             "https//github.com/dne/major-update",
			wantStderr: heredoc.Doc(`
				A new release of major-update-pin is available: 1.0.0 → 2.0.0
				To upgrade, run: gh extension upgrade major-update-pin --force
				https//github.com/dne/major-update
			`),
		},
		{
			name:               "minor update",
			extName:            "minor-update",
			extFullName:        "gh-minor-update",
			extUpdateAvailable: true,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "1.1.0",
			extURL:             "https//github.com/dne/minor-update",
			wantStderr: heredoc.Doc(`
				A new release of minor-update is available: 1.0.0 → 1.1.0
				To upgrade, run: gh extension upgrade minor-update
				https//github.com/dne/minor-update
			`),
		},
		{
			name:               "minor update, pinned",
			extName:            "minor-update-pin",
			extFullName:        "gh-minor-update-pin",
			extUpdateAvailable: true,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "1.1.0",
			extURL:             "https//github.com/dne/minor-update",
			extIsPinned:        true,
			wantStderr: heredoc.Doc(`
				A new release of minor-update-pin is available: 1.0.0 → 1.1.0
				To upgrade, run: gh extension upgrade minor-update-pin --force
				https//github.com/dne/minor-update
			`),
		},
		{
			name:               "patch update",
			extName:            "patch-update",
			extFullName:        "gh-patch-update",
			extUpdateAvailable: true,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "1.0.1",
			extURL:             "https//github.com/dne/patch-update",
			wantStderr: heredoc.Doc(`
				A new release of patch-update is available: 1.0.0 → 1.0.1
				To upgrade, run: gh extension upgrade patch-update
				https//github.com/dne/patch-update
			`),
		},
		{
			name:               "patch update, pinned",
			extName:            "patch-update-pin",
			extFullName:        "gh-patch-update-pin",
			extUpdateAvailable: true,
			extCurrentVersion:  "1.0.0",
			extLatestVersion:   "1.0.1",
			extURL:             "https//github.com/dne/patch-update",
			extIsPinned:        true,
			wantStderr: heredoc.Doc(`
				A new release of patch-update-pin is available: 1.0.0 → 1.0.1
				To upgrade, run: gh extension upgrade patch-update-pin --force
				https//github.com/dne/patch-update
			`),
		},
	}

	for _, tt := range tests {
		ios, _, _, stderr := iostreams.Test()

		em := &extensions.ExtensionManagerMock{
			DispatchFunc: func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error) {
				// Assume extension executed / dispatched without problems as test is focused on upgrade checking.
				return true, nil
			},
		}

		ext := &extensions.ExtensionMock{
			CurrentVersionFunc: func() string {
				return tt.extCurrentVersion
			},
			FullNameFunc: func() string {
				return tt.extFullName
			},
			IsPinnedFunc: func() bool {
				return tt.extIsPinned
			},
			LatestVersionFunc: func() string {
				return tt.extLatestVersion
			},
			NameFunc: func() string {
				return tt.extName
			},
			UpdateAvailableFunc: func() bool {
				return tt.extUpdateAvailable
			},
			URLFunc: func() string {
				return tt.extURL
			},
		}

		checkFunc := func(em extensions.ExtensionManager, ext extensions.Extension) (*update.ReleaseInfo, error) {
			if !tt.extUpdateAvailable {
				return nil, nil
			}

			return &update.ReleaseInfo{
				Version: tt.extLatestVersion,
				URL:     tt.extURL,
			}, nil
		}

		cmd := root.NewCmdExtension(ios, em, ext, checkFunc)

		_, err := cmd.ExecuteC()
		require.NoError(t, err)

		if tt.wantStderr == "" {
			assert.Emptyf(t, stderr.String(), "executing extension command should output nothing to stderr")
		} else {
			assert.Containsf(t, stderr.String(), tt.wantStderr, "executing extension command should output message about upgrade to stderr")
		}
	}
}
