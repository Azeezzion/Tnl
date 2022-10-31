package add

import (
	"net/http"
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func Test_runAdd(t *testing.T) {
	tests := []struct {
		name       string
		stdin      string
		httpStubs  func(*httpmock.Registry)
		wantsErr   bool
		wantErrMsg string
		wantsOpts  AddOptions
	}{
		{"armored_valid", "-----BEGIN PGP PUBLIC KEY BLOCK-----", func(reg *httpmock.Registry) {
			reg.Register(
				httpmock.REST("POST", "user/gpg_keys"),
				httpmock.WithHeader(
					httpmock.StatusStringResponse(200, `{}`),
					"Content-Type",
					"application/json",
				),
			)
		}, false, "", AddOptions{}},
		{"not_armored", "gCAAAAA7H7MHTZWFLJKD3vP4F7v", func(reg *httpmock.Registry) {
			reg.Register(
				httpmock.REST("POST", "user/gpg_keys"),
				httpmock.WithHeader(
					httpmock.StatusStringResponse(422, `{
                                                "message": "Validation Failed",
                                                "errors": [{
                                                        "resource": "GpgKey",
                                                        "code": "custom",
                                                        "message": "We got an error doing that."
                                                }],
                                                "documentation_url": "https://docs.github.com/v3/users/gpg_keys"
                                        }`),
					"Content-Type",
					"application/json",
				),
			)
		}, true, "it seems that the GPG key is not armored.\nplease try to find your GPG key ID using:\n\tgpg --list-keys\n" +
			"and use command below to add it to your accont:\n\tgpg --armor --export <GPG key ID> | gh gpg-key add -", AddOptions{}},
		{"duplicate", "-----BEGIN PGP PUBLIC KEY BLOCK-----", func(reg *httpmock.Registry) {
			reg.Register(
				httpmock.REST("POST", "user/gpg_keys"),
				httpmock.WithHeader(
					httpmock.StatusStringResponse(422, `{
                                                "message": "Validation Failed",
                                                "errors": [{
                                                        "resource": "GpgKey",
                                                        "code": "custom",
                                                        "field": "key_id",
                                                        "message": "key_id already exists"
                                                }, {
                                                        "resource": "GpgKey",
                                                        "code": "custom",
                                                        "field": "public_key",
                                                        "message": "public_key already exists"
                                                }],
                                                "documentation_url": "https://docs.github.com/v3/users/gpg_keys"
                                        }`),
					"Content-Type",
					"application/json",
				),
			)
		}, false, "", AddOptions{}},
	}

	for _, tt := range tests {
		ios, stdin, _, _ := iostreams.Test()
		reg := &httpmock.Registry{}
		defer reg.Verify(t)

		tt.wantsOpts.IO = ios
		tt.wantsOpts.HTTPClient = func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		}
		if tt.httpStubs != nil {
			tt.httpStubs(reg)
		}
		tt.wantsOpts.Config = func() (config.Config, error) {
			return config.NewBlankConfig(), nil
		}
		tt.wantsOpts.KeyFile = "-"

		stdin.WriteString(tt.stdin)
		t.Run(tt.name, func(t *testing.T) {
			err := runAdd(&tt.wantsOpts)
			if tt.wantsErr {
				if assert.Error(t, err) {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
