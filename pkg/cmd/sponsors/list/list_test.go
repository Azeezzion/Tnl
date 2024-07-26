package list

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/cli/v2/internal/prompter"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdList(t *testing.T) {
	tests := []struct {
		name          string
		cli           string
		prompterStubs func(*prompter.PrompterMock)
		wants         string
	}{
		{
			name:          "happy path",
			cli:           "octocat",
			prompterStubs: func(p *prompter.PrompterMock) {},
			wants:         "octocat",
		},
		{
			name: "no arguments",
			cli:  "",
			prompterStubs: func(p *prompter.PrompterMock) {
				p.InputFunc = func(p, d string) (string, error) {
					switch p {
					case "Which user do you want to target?":
						return "octocat", nil
					default:
						return d, nil
					}
				}
			},
			wants: "octocat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			pm := &prompter.PrompterMock{}
			if tt.prompterStubs != nil {
				tt.prompterStubs(pm)
			}
			f := &cmdutil.Factory{
				IOStreams: ios,
				Prompter:  pm,
			}

			argv, err := shlex.Split(tt.cli)
			assert.NoError(t, err)

			var gotOpts *ListOptions
			cmd := NewCmdList(f, func(opts *ListOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			assert.NoError(t, err)

			assert.Equal(t, tt.wants, gotOpts.Username)
		})
	}
}

func Test_listRun(t *testing.T) {
	tests := []struct {
		name  string
		opts  *ListOptions
		wants []string
	}{
		{
			name: "happy path",
			opts: &ListOptions{
				HttpClient: func() (*http.Client, error) {
					r := &httpmock.Registry{}
					return &http.Client{Transport: r}, nil
				},
				Getter:   NewSponsorsListGetter(getterFactory([]string{"mona", "lisa"}, nil)),
				Username: "octocat",
				Sponsors: []string{},
			},
			wants: []string{"mona", "lisa"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			ios.SetStdoutTTY(true)
			tt.opts.IO = ios

			err := listRun(tt.opts)
			assert.NoError(t, err)
			assert.Equal(t, tt.wants, tt.opts.Sponsors)
		})
	}
}

func Test_formatOutput(t *testing.T) {
	tests := []struct {
		name  string
		opts  *ListOptions
		wants []string
		isTTY bool
	}{
		{
			name: "Simple table",
			opts: &ListOptions{
				Username: "octocat",
				Sponsors: []string{"mona", "lisa"},
			},
			wants: []string{
				"SPONSORS",
				"mona",
				"lisa",
				"",
			},
			isTTY: true,
		},
		{
			name: "No Sponsors for octocat",
			opts: &ListOptions{
				Username: "octocat",
				Sponsors: []string{},
			},
			wants: []string{"No sponsors found for octocat\n"},
			isTTY: true,
		},
		{
			name: "No Sponsors for monalisa",
			opts: &ListOptions{
				Username: "monalisa",
				Sponsors: []string{},
			},
			wants: []string{"No sponsors found for monalisa\n"},
			isTTY: true,
		},
		{
			name: "No Sponsors for octocat non-TTY",
			opts: &ListOptions{
				Username: "octocat",
				Sponsors: []string{},
			},
			wants: []string{},
			isTTY: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, _ := iostreams.Test()
			ios.SetStdoutTTY(tt.isTTY)
			tt.opts.IO = ios

			err := formatOutput(tt.opts)
			assert.NoError(t, err)

			expected := strings.Join(tt.wants, "\n")

			assert.Equal(t, expected, stdout.String())
		})
	}
}
