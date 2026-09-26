package sandbox

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ai-tool-collection/WeKnora/internal/types"
)

// Identity comes from the stored row alone. Nothing about the deployment can
// change it, which is what makes "did this edit strand a sandbox" answerable
// without reaching for process-level sandbox configuration.
func TestIdentityOfReadsStoredFieldsOnly(t *testing.T) {
	identity := IdentityOf(&types.TenantSandboxConfig{
		SandboxType: "e2b",
		E2B: &types.E2BSandboxConfig{
			APIURL: "https://api.e2b.app", SandboxDomain: "e2b.app", APIKey: "key-a",
		},
	})

	require.Equal(t, SandboxIdentity{
		Provider:      "e2b",
		APIURL:        "https://api.e2b.app",
		APIKey:        "key-a",
		SandboxDomain: "e2b.app",
	}, identity)
}

// A blank field is blank, not "whatever the deployment runs". Two configs that
// differ only in what they spell out are therefore different identities.
func TestIdentityOfDockerWithoutHostCarriesProviderOnly(t *testing.T) {
	docker := IdentityOf(&types.TenantSandboxConfig{SandboxType: "docker"})
	cube := IdentityOf(&types.TenantSandboxConfig{SandboxType: "cube"})

	require.Equal(t, SandboxIdentity{Provider: "docker"}, docker)
	require.NotEqual(t, docker, cube)
}

// A nil config must not panic: Update compares against it when creating.
func TestIdentityOfToleratesNilConfig(t *testing.T) {
	require.Equal(t, SandboxIdentity{}, IdentityOf(nil))
}
