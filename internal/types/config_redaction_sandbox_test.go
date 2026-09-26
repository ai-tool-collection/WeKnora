package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSandboxConfigForResponseSkipsMaskingWhenDisabled(t *testing.T) {
	cfg := &TenantSandboxConfig{E2B: &E2BSandboxConfig{APIKey: "e2b-secret"}}

	out := SandboxConfigForResponse(cfg, false)

	require.Equal(t, "e2b-secret", out.E2B.APIKey)
}

func TestSandboxConfigForResponseEmptySecretStaysEmpty(t *testing.T) {
	cfg := &TenantSandboxConfig{E2B: &E2BSandboxConfig{APIKey: ""}}

	out := SandboxConfigForResponse(cfg, true)

	require.Empty(t, out.E2B.APIKey, "an unset secret must not become a placeholder")
}

func TestSandboxConfigForResponseNil(t *testing.T) {
	require.Nil(t, SandboxConfigForResponse(nil, true))
}

func TestMergeSandboxConfigForUpdateClearsPublicInbound(t *testing.T) {
	incoming := &TenantSandboxConfig{
		Network: &SandboxNetworkPolicy{AllowPublicInbound: true, AllowOut: []string{"1.1.1.1"}},
	}

	out := MergeSandboxConfigForUpdate(incoming, nil)

	require.NotNil(t, out.Network)
	require.False(t, out.Network.AllowPublicInbound,
		"inbound cannot be opened from a saved payload; the field is accepted then cleared")
	require.Equal(t, []string{"1.1.1.1"}, out.Network.AllowOut)
}

func TestMergeSandboxConfigForUpdateDropsNetworkWhenIncomingOmitsIt(t *testing.T) {
	existing := &TenantSandboxConfig{
		Network: &SandboxNetworkPolicy{DenyEgressByDefault: true},
	}

	out := MergeSandboxConfigForUpdate(&TenantSandboxConfig{}, existing)

	require.Nil(t, out.Network,
		"network is an editor-owned field: omitting it must clear it, unlike SkillImage")
}

func TestMergeSandboxConfigForUpdateHandlesNilExisting(t *testing.T) {
	incoming := &TenantSandboxConfig{E2B: &E2BSandboxConfig{APIKey: RedactedSecretPlaceholder}}

	out := MergeSandboxConfigForUpdate(incoming, nil)

	require.Empty(t, out.E2B.APIKey, "placeholder with no stored value resolves to empty")
}

func TestMergeSandboxConfigForUpdateNilIncoming(t *testing.T) {
	require.Nil(t, MergeSandboxConfigForUpdate(nil, &TenantSandboxConfig{}))
}

func TestMergeSandboxConfigForUpdatePreservesDesktopEnabled(t *testing.T) {
	incoming := &TenantSandboxConfig{SandboxType: "cube", DesktopEnabled: true}
	existing := &TenantSandboxConfig{SandboxType: "cube"}
	merged := MergeSandboxConfigForUpdate(incoming, existing)
	require.True(t, merged.DesktopEnabled)
}

func TestMergeSandboxConfigForUpdateClearsDesktopEnabled(t *testing.T) {
	incoming := &TenantSandboxConfig{SandboxType: "cube", DesktopEnabled: false}
	existing := &TenantSandboxConfig{SandboxType: "cube", DesktopEnabled: true}
	merged := MergeSandboxConfigForUpdate(incoming, existing)
	require.False(t, merged.DesktopEnabled,
		"selecting the CLI template must persist desktop_enabled=false, not keep a stale true")
}

func TestMergeSandboxConfigForUpdatePreservesSkillImage(t *testing.T) {
	existing := &TenantSandboxConfig{
		E2B:        &E2BSandboxConfig{APIKey: "old-e2b"},
		SkillImage: &SkillImageConfig{SnapshotID: "snap-1", OwnerFingerprint: "fp-1", Generation: 3},
	}
	incoming := &TenantSandboxConfig{
		E2B:        &E2BSandboxConfig{APIKey: RedactedSecretPlaceholder},
		SkillImage: &SkillImageConfig{SnapshotID: "forged-snap", OwnerFingerprint: "forged-fp"},
	}

	out := MergeSandboxConfigForUpdate(incoming, existing)

	require.Equal(t, "snap-1", out.SkillImage.SnapshotID,
		"a settings save must not replace the install-owned snapshot pointer")
	require.Equal(t, "fp-1", out.SkillImage.OwnerFingerprint)
	require.Equal(t, 3, out.SkillImage.Generation)
	out.SkillImage.SnapshotID = "mutated"
	require.Equal(t, "snap-1", existing.SkillImage.SnapshotID,
		"merge must copy SkillImage so later mutation cannot touch the stored row")
}

func TestMergeSandboxConfigForUpdateIgnoresIncomingSkillImageOnCreate(t *testing.T) {
	incoming := &TenantSandboxConfig{
		E2B:        &E2BSandboxConfig{APIKey: "new-e2b"},
		SkillImage: &SkillImageConfig{SnapshotID: "forged-snap"},
	}

	out := MergeSandboxConfigForUpdate(incoming, nil)

	require.Nil(t, out.SkillImage, "create must not accept a client-supplied skill image")
}

// The editor never sends skill_image: that pointer is written only by an
// install or a removal. A merge that copied the incoming payload as-is would
// therefore wipe a live snapshot on every "保存运行配置", leaving the skill
// rows in place while every session fell back to the base template. VolumeMount
// is covered here too because the editor omits it for the same reason.
func TestMergeSandboxConfigForUpdateHonoursExplicitSkillRollout(t *testing.T) {
	existing := &TenantSandboxConfig{SkillRollout: SkillRolloutNewSession}
	incoming := &TenantSandboxConfig{SkillRollout: SkillRolloutNextTurn}

	out := MergeSandboxConfigForUpdate(incoming, existing)

	require.Equal(t, SkillRolloutNextTurn, out.SkillRollout)
}

func TestMergeSandboxConfigForUpdateDoesNotMutateInputs(t *testing.T) {
	existing := &TenantSandboxConfig{E2B: &E2BSandboxConfig{APIKey: "old-e2b"}}
	incoming := &TenantSandboxConfig{
		E2B:     &E2BSandboxConfig{APIKey: RedactedSecretPlaceholder},
		EnvVars: map[string]string{"HF_TOKEN": RedactedSecretPlaceholder},
	}

	_ = MergeSandboxConfigForUpdate(incoming, existing)

	require.Equal(t, RedactedSecretPlaceholder, incoming.E2B.APIKey,
		"merge must not mutate the incoming payload")
	require.Equal(t, "old-e2b", existing.E2B.APIKey,
		"merge must not mutate the stored config")
}
