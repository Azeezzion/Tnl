package verify

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cli/cli/v2/internal/ghinstance"
	"github.com/cli/cli/v2/pkg/cmdutil"
)

// Options captures the options for the verify command
type Options struct {
	ArtifactPath         string
	BundlePath           string
	DenySelfHostedRunner bool
	DigestAlgorithm      string
	exporter             cmdutil.Exporter
	Hostname             string
	Limit                int
	NoPublicGood         bool
	OIDCIssuer           string
	Owner                string
	PredicateType        string
	Repo                 string
	SAN                  string
	SANRegex             string
	SignerRepo           string
	SignerWorkflow       string
	// Tenant is only set when tenancy is used
	Tenant                string
	TrustedRoot           string
	UseBundleFromRegistry bool
}

// Clean cleans the file path option values
func (opts *Options) Clean() {
	if opts.BundlePath != "" {
		opts.BundlePath = filepath.Clean(opts.BundlePath)
	}
}

func (opts *Options) SetPolicyFlags() {
	// check that Repo is in the expected format if provided
	if opts.Repo != "" {
		// we expect the repo argument to be in the format <OWNER>/<REPO>
		splitRepo := strings.Split(opts.Repo, "/")

		// if Repo is provided but owner is not, set the OWNER portion of the Repo value
		// to Owner
		opts.Owner = splitRepo[0]

		if !isSignerIdentityProvided(opts) {
			opts.SANRegex = expandToGitHubURL(opts.Tenant, opts.Repo)
		}
		return
	}
	if !isSignerIdentityProvided(opts) {
		opts.SANRegex = expandToGitHubURL(opts.Tenant, opts.Owner)
	}
}

// AreFlagsValid checks that the provided flag combination is valid
// and returns an error otherwise
func (opts *Options) AreFlagsValid() error {
	// If provided, check that the Repo option is in the expected format <OWNER>/<REPO>
	if opts.Repo != "" && !isProvidedRepoValid(opts.Repo) {
		return fmt.Errorf("invalid value provided for repo: %s", opts.Repo)
	}

	// If provided, check that the SignerRepo option is in the expected format <OWNER>/<REPO>
	if opts.SignerRepo != "" && !isProvidedRepoValid(opts.SignerRepo) {
		return fmt.Errorf("invalid value provided for signer-repo: %s", opts.SignerRepo)
	}

	// Check that limit is between 1 and 1000
	if opts.Limit < 1 || opts.Limit > 1000 {
		return fmt.Errorf("limit %d not allowed, must be between 1 and 1000", opts.Limit)
	}

	// Check that the bundle-from-oci flag is only used with OCI artifact paths
	if opts.UseBundleFromRegistry && !strings.HasPrefix(opts.ArtifactPath, "oci://") {
		return fmt.Errorf("bundle-from-oci flag can only be used with OCI artifact paths")
	}

	// Check that both the bundle-from-oci and bundle-path flags are not used together
	if opts.UseBundleFromRegistry && opts.BundlePath != "" {
		return fmt.Errorf("bundle-from-oci flag cannot be used with bundle-path flag")
	}

	// Verify provided hostname
	if opts.Hostname != "" {
		if err := ghinstance.HostnameValidator(opts.Hostname); err != nil {
			return fmt.Errorf("error parsing hostname: %w", err)
		}
	}

	return nil
}

// check if any of the signer identity flags have been provided
func isSignerIdentityProvided(opts *Options) bool {
	return opts.SAN != "" || opts.SANRegex != "" || opts.SignerRepo != "" || opts.SignerWorkflow != ""
}

func isProvidedRepoValid(repo string) bool {
	// we expect a provided repository argument be in the format <OWNER>/<REPO>
	splitRepo := strings.Split(repo, "/")
	return len(splitRepo) == 2
}
