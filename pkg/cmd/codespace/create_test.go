package codespace

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cli/cli/v2/internal/codespaces"
	"github.com/cli/cli/v2/internal/codespaces/api"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/cli/cli/v2/pkg/liveshare"
	"github.com/stretchr/testify/assert"
)

func TestApp_Create(t *testing.T) {
	type fields struct {
		apiClient apiClient
	}
	tests := []struct {
		name       string
		fields     fields
		opts       createOptions
		wantErr    error
		wantStdout string
		wantStderr string
		isTTY      bool
	}{
		{
			name: "create codespace with default branch and 30m idle timeout",
			fields: fields{
				apiClient: &apiClientMock{
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            1234,
							FullName:      nwo,
							DefaultBranch: "main",
						}, nil
					},
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return []api.DevContainerEntry{{Path: ".devcontainer/devcontainer.json"}}, nil
					},
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string) ([]*api.Machine, error) {
						return []*api.Machine{
							{
								Name:        "GIGA",
								DisplayName: "Gigabits of a machine",
							},
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf("got branch %q, want %q", params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf("idle timeout minutes was %v", params.IdleTimeoutMinutes)
						}
						if *params.RetentionPeriodMinutes != 2880 {
							return nil, fmt.Errorf("retention period minutes expected 2880, was %v", params.RetentionPeriodMinutes)
						}
						return &api.Codespace{
							Name: "monalisa-dotfiles-abcd1234",
						}, nil
					},
					GetCodespaceRepoSuggestionsFunc: func(ctx context.Context, partialSearch string, params api.RepoSearchParameters) ([]string, error) {
						return nil, nil // We can't ask for suggestions without a terminal.
					},
				},
			},
			opts: createOptions{
				repo:            "monalisa/dotfiles",
				branch:          "",
				machine:         "GIGA",
				showStatus:      false,
				idleTimeout:     30 * time.Minute,
				retentionPeriod: NullableDuration{durationPtr(48 * time.Hour)},
			},
			wantStdout: "monalisa-dotfiles-abcd1234\n",
		},
		{
			name: "create codespace with default branch shows idle timeout notice if present",
			fields: fields{
				apiClient: &apiClientMock{
					GetCodespaceRegionLocationFunc: func(ctx context.Context) (string, error) {
						return "EUROPE", nil
					},
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            1234,
							FullName:      nwo,
							DefaultBranch: "main",
						}, nil
					},
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string) ([]*api.Machine, error) {
						return []*api.Machine{
							{
								Name:        "GIGA",
								DisplayName: "Gigabits of a machine",
							},
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf("got branch %q, want %q", params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf("idle timeout minutes was %v", params.IdleTimeoutMinutes)
						}
						if params.RetentionPeriodMinutes != nil {
							return nil, fmt.Errorf("retention period minutes expected nil, was %v", params.RetentionPeriodMinutes)
						}
						if params.DevContainerPath != ".devcontainer/foobar/devcontainer.json" {
							return nil, fmt.Errorf("got dev container path %q, want %q", params.DevContainerPath, ".devcontainer/foobar/devcontainer.json")
						}
						return &api.Codespace{
							Name: "monalisa-dotfiles-abcd1234",
						}, nil
					},
				},
			},
			opts: createOptions{
				repo:             "monalisa/dotfiles",
				branch:           "",
				machine:          "GIGA",
				showStatus:       false,
				idleTimeout:      30 * time.Minute,
				devContainerPath: ".devcontainer/foobar/devcontainer.json",
			},
			wantStdout: "monalisa-dotfiles-abcd1234\n",
		},
		{
			name: "create codespace with default branch with default devcontainer if no path provided and no devcontainer files exist in the repo",
			fields: fields{
				apiClient: &apiClientMock{
					GetCodespaceRegionLocationFunc: func(ctx context.Context) (string, error) {
						return "EUROPE", nil
					},
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            1234,
							FullName:      nwo,
							DefaultBranch: "main",
						}, nil
					},
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return []api.DevContainerEntry{}, nil
					},
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string) ([]*api.Machine, error) {
						return []*api.Machine{
							{
								Name:        "GIGA",
								DisplayName: "Gigabits of a machine",
							},
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf("got branch %q, want %q", params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf("idle timeout minutes was %v", params.IdleTimeoutMinutes)
						}
						if params.DevContainerPath != "" {
							return nil, fmt.Errorf("got dev container path %q, want %q", params.DevContainerPath, ".devcontainer/foobar/devcontainer.json")
						}
						return &api.Codespace{
							Name:              "monalisa-dotfiles-abcd1234",
							IdleTimeoutNotice: "Idle timeout for this codespace is set to 10 minutes in compliance with your organization's policy",
						}, nil
					},
					GetCodespaceRepoSuggestionsFunc: func(ctx context.Context, partialSearch string, params api.RepoSearchParameters) ([]string, error) {
						return nil, nil // We can't ask for suggestions without a terminal.
					},
				},
			},
			opts: createOptions{
				repo:        "monalisa/dotfiles",
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantStdout: "monalisa-dotfiles-abcd1234\n",
			wantStderr: "Notice: Idle timeout for this codespace is set to 10 minutes in compliance with your organization's policy\n",
			isTTY:      true,
		},
		{
			name: "returns error when getting devcontainer paths fails",
			fields: fields{
				apiClient: &apiClientMock{
					GetCodespaceRegionLocationFunc: func(ctx context.Context) (string, error) {
						return "EUROPE", nil
					},
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            1234,
							FullName:      nwo,
							DefaultBranch: "main",
						}, nil
					},
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return nil, fmt.Errorf("some error")
					},
				},
			},
			opts: createOptions{
				repo:        "monalisa/dotfiles",
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantErr: fmt.Errorf("error getting devcontainer.json paths: some error"),
		},
		{
			name: "create codespace with default branch does not show idle timeout notice if not conntected to terminal",
			fields: fields{
				apiClient: &apiClientMock{
					GetCodespaceRegionLocationFunc: func(ctx context.Context) (string, error) {
						return "EUROPE", nil
					},
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            1234,
							FullName:      nwo,
							DefaultBranch: "main",
						}, nil
					},
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return []api.DevContainerEntry{}, nil
					},
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string) ([]*api.Machine, error) {
						return []*api.Machine{
							{
								Name:        "GIGA",
								DisplayName: "Gigabits of a machine",
							},
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf("got branch %q, want %q", params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf("idle timeout minutes was %v", params.IdleTimeoutMinutes)
						}
						return &api.Codespace{
							Name:              "monalisa-dotfiles-abcd1234",
							IdleTimeoutNotice: "Idle timeout for this codespace is set to 10 minutes in compliance with your organization's policy",
						}, nil
					},
					GetCodespaceRepoSuggestionsFunc: func(ctx context.Context, partialSearch string, params api.RepoSearchParameters) ([]string, error) {
						return nil, nil // We can't ask for suggestions without a terminal.
					},
				},
			},
			opts: createOptions{
				repo:        "monalisa/dotfiles",
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantStdout: "monalisa-dotfiles-abcd1234\n",
			wantStderr: "",
			isTTY:      false,
		},
		{
			name: "create codespace that requires accepting additional permissions",
			fields: fields{
				apiClient: &apiClientMock{
					GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
						return &api.Repository{
							ID:            1234,
							FullName:      nwo,
							DefaultBranch: "main",
						}, nil
					},
					ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
						return []api.DevContainerEntry{{Path: ".devcontainer/devcontainer.json"}}, nil
					},
					GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string) ([]*api.Machine, error) {
						return []*api.Machine{
							{
								Name:        "GIGA",
								DisplayName: "Gigabits of a machine",
							},
						}, nil
					},
					CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
						if params.Branch != "main" {
							return nil, fmt.Errorf("got branch %q, want %q", params.Branch, "main")
						}
						if params.IdleTimeoutMinutes != 30 {
							return nil, fmt.Errorf("idle timeout minutes was %v", params.IdleTimeoutMinutes)
						}
						return &api.Codespace{}, api.AcceptPermissionsRequiredError{
							AllowPermissionsURL: "https://example.com/permissions",
						}
					},
					GetCodespaceRepoSuggestionsFunc: func(ctx context.Context, partialSearch string, params api.RepoSearchParameters) ([]string, error) {
						return nil, nil // We can't ask for suggestions without a terminal.
					},
				},
			},
			opts: createOptions{
				repo:        "monalisa/dotfiles",
				branch:      "",
				machine:     "GIGA",
				showStatus:  false,
				idleTimeout: 30 * time.Minute,
			},
			wantErr: cmdutil.SilentError,
			wantStderr: `You must authorize or deny additional permissions requested by this codespace before continuing.
Open this URL in your browser to review and authorize additional permissions: example.com/permissions
Alternatively, you can run "create" with the "--default-permissions" option to continue without authorizing additional permissions.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			ios.SetStdinTTY(tt.isTTY)
			ios.SetStderrTTY(tt.isTTY)

			a := &App{
				io:        ios,
				apiClient: tt.fields.apiClient,
			}

			err := a.Create(context.Background(), tt.opts)
			if err != nil && tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
			if err != nil && tt.wantErr == nil {
				t.Logf(err.Error())
			}
			if got := stdout.String(); got != tt.wantStdout {
				t.Logf(t.Name())
				t.Errorf("  stdout = %v, want %v", got, tt.wantStdout)
			}
			if got := stderr.String(); got != tt.wantStderr {
				t.Logf(t.Name())
				t.Errorf("  stderr = %v, want %v", got, tt.wantStderr)
			}
		})
	}
}

func TestBuildDisplayName(t *testing.T) {
	tests := []struct {
		name                 string
		prebuildAvailability string
		expectedDisplayName  string
	}{
		{
			name:                 "prebuild availability is pool",
			prebuildAvailability: "pool",
			expectedDisplayName:  "4 cores, 8 GB RAM, 32 GB storage (Prebuild ready)",
		},
		{
			name:                 "prebuild availability is blob",
			prebuildAvailability: "blob",
			expectedDisplayName:  "4 cores, 8 GB RAM, 32 GB storage (Prebuild ready)",
		},
		{
			name:                 "prebuild availability is none",
			prebuildAvailability: "none",
			expectedDisplayName:  "4 cores, 8 GB RAM, 32 GB storage",
		},
		{
			name:                 "prebuild availability is empty",
			prebuildAvailability: "",
			expectedDisplayName:  "4 cores, 8 GB RAM, 32 GB storage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			displayName := buildDisplayName("4 cores, 8 GB RAM, 32 GB storage", tt.prebuildAvailability)

			if displayName != tt.expectedDisplayName {
				t.Errorf("displayName = %q, expectedDisplayName %q", displayName, tt.expectedDisplayName)
			}
		})
	}
}

func TestCreateAndSsh(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	user := &api.User{Login: "monalisa"}
	apiMock := &apiClientMock{
		ListDevContainersFunc: func(ctx context.Context, repoID int, branch string, limit int) ([]api.DevContainerEntry, error) {
			return []api.DevContainerEntry{{Path: ".devcontainer/devcontainer.json"}}, nil
		},
		GetCodespacesMachinesFunc: func(ctx context.Context, repoID int, branch, location string) ([]*api.Machine, error) {
			return []*api.Machine{
				{
					Name:        "GIGA",
					DisplayName: "Gigabits of a machine",
				},
			}, nil
		},
		CreateCodespaceFunc: func(ctx context.Context, params *api.CreateCodespaceParams) (*api.Codespace, error) {
			return &api.Codespace{
				Name: "monalisa-dotfiles-abcd1234",
			}, nil
		},
		GetRepositoryFunc: func(ctx context.Context, nwo string) (*api.Repository, error) {
			return &api.Repository{
				ID:            1234,
				FullName:      "monalisa/dotfiles",
				DefaultBranch: "main",
			}, nil
		},
		GetUserFunc: func(_ context.Context) (*api.User, error) {
			return user, nil
		},
		AuthorizedKeysFunc: func(_ context.Context, _ string) ([]byte, error) {
			return []byte{}, nil
		},
		StartCodespaceFunc: func(ctx context.Context, name string) error {
			fmt.Println("Starting codespace")
			return nil
		},
		GetCodespaceFunc: func(ctx context.Context, name string, includeConnection bool) (*api.Codespace, error) {
			return &api.Codespace{
				Name:  "monalisa-dotfiles-abcd1234",
				State: api.CodespaceStateAvailable,
				Connection: api.CodespaceConnection{
					SessionID:     "something",
					SessionToken:  "something",
					RelayEndpoint: "something",
					RelaySAS:      "something",
				},
			}, nil
		},
	}

	a := &App{
		io:              ios,
		apiClient:       apiMock,
		liveshareClient: mockLiveshareClient{},
	}

	PollStates = func(ctx context.Context, progress codespaces.ProgressIndicator, apiClient codespaces.ApiClient, codespace *api.Codespace, poller func([]codespaces.PostCreateState)) (err error) {

		if codespace.Name != "monalisa-dotfiles-abcd1234" {
			t.Errorf("Expected codespace name to be 'monalisa-dotfiles-abcd1234'. Got %q", codespace.Name)
		}

		return nil
	}

	opts := createOptions{
		repo:             "monalisa/dotfiles",
		branch:           "",
		machine:          "GIGA",
		ssh:              true,
		showStatus:       false,
		idleTimeout:      30 * time.Minute,
		devContainerPath: ".devcontainer/foobar/devcontainer.json",
	}

	err := a.Create(context.Background(), opts)
	if err != nil {
		t.Error(err)
	}
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}

type mockLiveshareSession struct{}

func (s mockLiveshareSession) Close() error {
	return nil
}

func (s mockLiveshareSession) StartSSHServer(ctx context.Context) (int, string, error) {
	return 0, "", nil
}

func (s mockLiveshareSession) StartJupyterServer(ctx context.Context) (int, string, error) {
	return 0, "", nil
}

type mockLiveshareClient struct{}

func (client mockLiveshareClient) startLiveShareSession(ctx context.Context, codespace *api.Codespace, a *App, debug bool, debugFile string) (liveshare.LiveshareSession, error) {
	fmt.Println("WORKSSSSSS")
	return mockLiveshareSession{}, nil
}
