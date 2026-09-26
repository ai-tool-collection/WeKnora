package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// setAESKey installs a valid 32-byte SYSTEM_AES_KEY for the test.
func setAESKey(t *testing.T) {
	t.Helper()
	t.Setenv("SYSTEM_AES_KEY", strings.Repeat("k", 32))
}

func TestTenantSandboxConfigScanTolerantOfUndecryptableSecrets(t *testing.T) {
	// A row written under a different/rotated key must not break loading the
	// tenant; the secret is blanked and treated as unconfigured.
	setAESKey(t)

	payload, err := json.Marshal(map[string]any{
		"sandbox_type": "e2b",
		"e2b":          map[string]string{"api_key": "enc:v1:not-really-valid-base64!!"},
		"env_vars":     map[string]string{"HF_TOKEN": "enc:v1:also-invalid!!"},
	})
	require.NoError(t, err)

	var cfg TenantSandboxConfig
	require.NoError(t, cfg.Scan(payload), "Scan must not fail on undecryptable secrets")
	require.Empty(t, cfg.E2B.APIKey)
	require.Empty(t, cfg.EnvVars["HF_TOKEN"])
	require.Equal(t, "e2b", cfg.SandboxType, "non-secret fields still load")
}

func TestTenantSandboxConfigNilRoundTrip(t *testing.T) {
	var cfg *TenantSandboxConfig

	raw, err := cfg.Value()
	require.NoError(t, err)
	require.Nil(t, raw, "a nil config persists as SQL NULL")

	var restored TenantSandboxConfig
	require.NoError(t, restored.Scan(nil), "scanning NULL is a no-op")
}
