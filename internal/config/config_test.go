package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	ghConfig "github.com/cli/go-gh/v2/pkg/config"
)

func newTestConfig() *cfg {
	return &cfg{
		cfg: ghConfig.ReadFromString(""),
	}
}

func TestGetNonExistentKey(t *testing.T) {
	// Given we have no top level configuration
	cfg := newTestConfig()

	// When we get a key that has no value
	val, err := cfg.Get("", "non-existent-key")

	// Then it returns an error and the value is empty
	var keyNotFoundError *ghConfig.KeyNotFoundError
	require.ErrorAs(t, err, &keyNotFoundError)
	require.Empty(t, val)
}

func TestGetNonExistentHostSpecificKey(t *testing.T) {
	// Given have no top level configuration
	cfg := newTestConfig()

	// When we get a key for a host that has no value
	val, err := cfg.Get("non-existent-host", "non-existent-key")

	// Then it returns an error and the value is empty
	var keyNotFoundError *ghConfig.KeyNotFoundError
	require.ErrorAs(t, err, &keyNotFoundError)
	require.Empty(t, val)
}

func TestGetExistingTopLevelKey(t *testing.T) {
	// Given have a top level config entry
	cfg := newTestConfig()
	cfg.Set("", "top-level-key", "top-level-value")

	// When we get that key
	val, err := cfg.Get("non-existent-host", "top-level-key")

	// Then it returns successfully with the correct value
	require.NoError(t, err)
	require.Equal(t, "top-level-value", val)
}

func TestGetExistingHostSpecificKey(t *testing.T) {
	// Given have a host specific config entry
	cfg := newTestConfig()
	cfg.Set("github.com", "host-specific-key", "host-specific-value")

	// When we get that key
	val, err := cfg.Get("github.com", "host-specific-key")

	// Then it returns successfully with the correct value
	require.NoError(t, err)
	require.Equal(t, "host-specific-value", val)
}

func TestGetHostnameSpecificKeyFallsBackToTopLevel(t *testing.T) {
	// Given have a top level config entry
	cfg := newTestConfig()
	cfg.Set("", "key", "value")

	// When we get that key on a specific host
	val, err := cfg.Get("github.com", "key")

	// Then it returns successfully, falling back to the top level config
	require.NoError(t, err)
	require.Equal(t, "value", val)
}

func TestGetOrDefaultApplicationDefaults(t *testing.T) {
	tests := []struct {
		key             string
		expectedDefault string
	}{
		{"git_protocol", "https"},
		{"editor", ""},
		{"prompt", "enabled"},
		{"pager", ""},
		{"http_unix_socket", ""},
		{"browser", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			// Given we have no top level configuration
			cfg := newTestConfig()

			// When we get a key that has no value, but has a default
			val, err := cfg.GetOrDefault("", tt.key)

			// Then it returns the default value
			require.NoError(t, err)
			require.Equal(t, tt.expectedDefault, val)
		})
	}
}

func TestGetOrDefaultExistingKey(t *testing.T) {
	// Given have a top level config entry
	cfg := newTestConfig()
	cfg.Set("", "git_protocol", "ssh")

	// When we get that key
	val, err := cfg.GetOrDefault("", "git_protocol")

	// Then it returns successfully with the correct value, and doesn't fall back
	// to the default
	require.NoError(t, err)
	require.Equal(t, "ssh", val)
}

func TestGetOrDefaultNotFoundAndNoDefault(t *testing.T) {
	// Given have no configuration
	cfg := newTestConfig()

	// When we get a non-existent-key that has no default
	val, err := cfg.GetOrDefault("", "non-existent-key")

	// Then it returns an error and the value is empty
	var keyNotFoundError *ghConfig.KeyNotFoundError
	require.ErrorAs(t, err, &keyNotFoundError)
	require.Empty(t, val)
}

func TestFallbackConfig(t *testing.T) {
	cfg := fallbackConfig()

	protocol, err := cfg.Get([]string{"git_protocol"})
	require.NoError(t, err)
	require.Equal(t, "https", protocol)

	editor, err := cfg.Get([]string{"editor"})
	require.NoError(t, err)
	require.Equal(t, "", editor)

	prompt, err := cfg.Get([]string{"prompt"})
	require.NoError(t, err)
	require.Equal(t, "enabled", prompt)

	pager, err := cfg.Get([]string{"pager"})
	require.NoError(t, err)
	require.Equal(t, "", pager)

	socket, err := cfg.Get([]string{"http_unix_socket"})
	require.NoError(t, err)
	require.Equal(t, "", socket)

	browser, err := cfg.Get([]string{"browser"})
	require.NoError(t, err)
	require.Equal(t, "", browser)

	unknown, err := cfg.Get([]string{"unknown"})
	require.EqualError(t, err, `could not find key "unknown"`)
	require.Equal(t, "", unknown)
}
