package sandbox

import (
	"testing"
	"time"

	"github.com/ai-tool-collection/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

func globalTestConfig() *Config {
	cfg := &Config{
		Type:             SandboxTypeE2B,
		DefaultTimeout:   60 * time.Second,
		E2BAPIKey:        "global-key",
		E2BAPIURL:        "https://global.e2b.dev",
		E2BSandboxDomain: "global.domain",
		E2BTemplate:      "global-template",
		E2BSandboxTTL:    10 * time.Minute,
		E2BHTTPTimeout:   30 * time.Second,
	}
	cfg.Network = resolveNetworkPolicy(nil)
	return cfg
}

// completeE2BTenantConfig is the minimum a named E2B config must carry now that
// nothing is inherited.
func completeE2BTenantConfig() *types.TenantSandboxConfig {
	return &types.TenantSandboxConfig{
		SandboxType: "e2b",
		E2B: &types.E2BSandboxConfig{
			APIKey:     "tenant-key",
			TemplateID: "tenant-template",
		},
	}
}

func TestResolveEffectiveConfigNilTenantKeepsGlobal(t *testing.T) {
	global := globalTestConfig()

	got, err := ResolveEffectiveConfig(nil, global)

	require.NoError(t, err)
	require.Equal(t, *global, *got)
}

func TestResolveEffectiveConfigNilTenantMaterializesNetworkDefaults(t *testing.T) {
	global := &Config{Type: SandboxTypeDisabled, DefaultTimeout: time.Second}

	got, err := ResolveEffectiveConfig(nil, global)

	require.NoError(t, err)
	require.NotNil(t, got.Network.AllowInternetAccess)
	require.True(t, *got.Network.AllowInternetAccess)
	require.NotNil(t, got.Network.AllowPublicTraffic)
	require.False(t, *got.Network.AllowPublicTraffic)
}

func TestDefaultConfigMaterializesNetworkDefaults(t *testing.T) {
	got := DefaultConfig()

	require.NotNil(t, got.Network.AllowInternetAccess)
	require.True(t, *got.Network.AllowInternetAccess)
	require.NotNil(t, got.Network.AllowPublicTraffic)
	require.False(t, *got.Network.AllowPublicTraffic)
}

func TestResolveEffectiveConfigDoesNotMutateGlobal(t *testing.T) {
	global := globalTestConfig()

	_, err := ResolveEffectiveConfig(completeE2BTenantConfig(), global)

	require.NoError(t, err)
	require.Equal(t, "global-key", global.E2BAPIKey,
		"resolution must not leak tenant values into the shared global config")
}

// The point of the whole design: a named config never picks up the deployment's
// endpoint, domain or key, so what it does not state it does not get.
func TestResolveEffectiveConfigDoesNotInheritProviderFields(t *testing.T) {
	global := globalTestConfig()

	got, err := ResolveEffectiveConfig(completeE2BTenantConfig(), global)

	require.NoError(t, err)
	require.Equal(t, "tenant-key", got.E2BAPIKey)
	require.Equal(t, "tenant-template", got.E2BTemplate)
	require.Empty(t, got.E2BAPIURL, "go-e2b resolves its own API base when unset")
	require.Empty(t, got.E2BSandboxDomain)
}

func TestResolveEffectiveConfigTranslatesNetworkPolicy(t *testing.T) {
	tenantCfg := completeE2BTenantConfig()
	tenantCfg.Network = &types.SandboxNetworkPolicy{
		DenyEgressByDefault: true,
		AllowPublicInbound:  false,
		AllowOut:            []string{"*.example.com"},
		DenyOut:             []string{"0.0.0.0/0"},
		E2BHostRules: []types.E2BHostRule{{
			Host:    "api.example.com",
			Headers: map[string]string{"X-Key": "secret"},
		}},
	}

	got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())
	require.NoError(t, err)

	require.NotNil(t, got.Network.AllowInternetAccess)
	require.False(t, *got.Network.AllowInternetAccess,
		"DenyEgressByDefault must flip the provider's allow switch off")
	require.NotNil(t, got.Network.AllowPublicTraffic)
	require.False(t, *got.Network.AllowPublicTraffic,
		"inbound is always credential-required")
	require.Equal(t, []string{"*.example.com"}, got.Network.AllowOut)
	require.Equal(t, []string{"0.0.0.0/0"}, got.Network.DenyOut)
	require.Equal(t, "api.example.com", got.Network.E2BHostRules[0].Host)
	require.Equal(t, "secret", got.Network.E2BHostRules[0].Headers["X-Key"])
}

// E2B rejects a create whose allowOut names a domain unless denyOut carries an
// explicit deny-all; it does not accept allow_internet_access=false as a
// substitute. Since DenyEgressByDefault means exactly "install a 0.0.0.0/0
// deny-all", the resolver has to materialise that entry, or ticking the box the
// validator recommends produces a payload the provider refuses.
func TestResolveEffectiveConfigMaterialisesDenyAllEntry(t *testing.T) {
	tenantCfg := completeE2BTenantConfig()
	tenantCfg.Network = &types.SandboxNetworkPolicy{
		DenyEgressByDefault: true,
		AllowOut:            []string{"*.example.com"},
	}

	got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.NoError(t, err)
	require.Equal(t, []string{"0.0.0.0/0"}, got.Network.DenyOut)
	require.True(t, got.Network.DeniesEgressByDefault())
}

func TestResolveEffectiveConfigDoesNotDuplicateDenyAllEntry(t *testing.T) {
	for _, existing := range []string{"0.0.0.0/0", "  0.0.0.0/0  ", "1.2.3.4/0"} {
		t.Run(existing, func(t *testing.T) {
			tenantCfg := completeE2BTenantConfig()
			tenantCfg.Network = &types.SandboxNetworkPolicy{
				DenyEgressByDefault: true,
				AllowOut:            []string{"*.example.com"},
				DenyOut:             []string{existing},
			}

			got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

			require.NoError(t, err)
			require.Equal(t, []string{types.DenyAllIPv4}, got.Network.DenyOut,
				"any IPv4 /0 must be rewritten to the canonical deny-all, not duplicated")
		})
	}
}

func TestResolveEffectiveConfigCanonicalizesNonCanonicalDenyAll(t *testing.T) {
	tenantCfg := completeE2BTenantConfig()
	tenantCfg.Network = &types.SandboxNetworkPolicy{
		AllowOut: []string{"api.example.com"},
		DenyOut:  []string{"10.0.0.0/8", "1.2.3.4/0"},
	}

	got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.NoError(t, err)
	require.Equal(t, []string{"10.0.0.0/8", types.DenyAllIPv4}, got.Network.DenyOut)
	require.True(t, got.Network.DeniesEgressByDefault())
}

// The entry is the resolved form of DenyEgressByDefault, so an egress-open
// policy must not acquire it.
func TestResolveEffectiveConfigLeavesDenyOutAloneWhenEgressOpen(t *testing.T) {
	tenantCfg := completeE2BTenantConfig()
	tenantCfg.Network = &types.SandboxNetworkPolicy{
		DenyOut: []string{"169.254.169.254/32"},
	}

	got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.NoError(t, err)
	require.Equal(t, []string{"169.254.169.254/32"}, got.Network.DenyOut)
	require.False(t, got.Network.DeniesEgressByDefault())
}

func TestResolveEffectiveConfigIgnoresStoredPublicInbound(t *testing.T) {
	tenantCfg := completeE2BTenantConfig()
	tenantCfg.Network = &types.SandboxNetworkPolicy{AllowPublicInbound: true}

	got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.NoError(t, err)
	require.NotNil(t, got.Network.AllowPublicTraffic)
	require.False(t, *got.Network.AllowPublicTraffic,
		"inbound stays credential-required even if a stored row asked for public access")
}

func TestResolveEffectiveConfigDefaultsNetworkToEgressOpenInboundClosed(t *testing.T) {
	// No stored policy at all: the shipped default must still be explicit, so
	// the adapters do not have to re-derive it.
	got, err := ResolveEffectiveConfig(completeE2BTenantConfig(), globalTestConfig())
	require.NoError(t, err)

	require.NotNil(t, got.Network.AllowInternetAccess)
	require.True(t, *got.Network.AllowInternetAccess)
	require.NotNil(t, got.Network.AllowPublicTraffic)
	require.False(t, *got.Network.AllowPublicTraffic)
}

func TestResolveEffectiveConfigDoesNotInheritNetworkFromBaseline(t *testing.T) {
	// This assertion pins the resolved end state, not clearProviderFields as
	// the mechanism that produces it.
	baseline := globalTestConfig()
	inherited := true
	baseline.Network = RemoteNetworkPolicy{AllowPublicTraffic: &inherited}

	got, err := ResolveEffectiveConfig(completeE2BTenantConfig(), baseline)
	require.NoError(t, err)

	require.NotNil(t, got.Network.AllowPublicTraffic)
	require.False(t, *got.Network.AllowPublicTraffic,
		"policy belongs to the named config; the .env baseline must not leak in")
}

// A leftover sub-struct from the deployment's other provider must not survive
// either, or a cube config would silently answer with e2b coordinates.
func TestResolveEffectiveConfigRejectsIncompleteE2B(t *testing.T) {
	_, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "e2b",
		E2B:         &types.E2BSandboxConfig{TemplateID: "t1"},
	}, globalTestConfig())

	require.ErrorIs(t, err, ErrSandboxConfigIncomplete)
	require.Contains(t, err.Error(), "api_key")
}

func TestResolveEffectiveConfigAppliesTimeoutsAndTTL(t *testing.T) {
	global := globalTestConfig()
	tenantCfg := completeE2BTenantConfig()
	tenantCfg.DefaultTimeoutSec = 90
	tenantCfg.E2B.HTTPTimeoutSec = 15
	tenantCfg.E2B.E2BSandboxTTLSeconds = 600
	tenantCfg.TerminalIdleDisconnectSec = 1200

	got, err := ResolveEffectiveConfig(tenantCfg, global)

	require.NoError(t, err)
	require.Equal(t, 90*time.Second, got.DefaultTimeout)
	require.Equal(t, 15*time.Second, got.E2BHTTPTimeout)
	require.Equal(t, 600*time.Second, got.E2BSandboxTTL)
	require.Equal(t, 20*time.Minute, got.TerminalIdleDisconnect)
}

func TestResolveEffectiveConfigTerminalIdleFallsBackToBuiltIn(t *testing.T) {
	global := globalTestConfig()
	global.TerminalIdleDisconnect = 3 * time.Hour

	got, err := ResolveEffectiveConfig(completeE2BTenantConfig(), global)

	require.NoError(t, err)
	require.Equal(t, DefaultTerminalIdleDisconnect, got.TerminalIdleDisconnect)
}

func TestResolveEffectiveConfigTuningFallsBackToBuiltIns(t *testing.T) {
	global := globalTestConfig()
	global.E2BSandboxTTL = 10 * time.Minute
	global.E2BHTTPTimeout = 90 * time.Second

	got, err := ResolveEffectiveConfig(completeE2BTenantConfig(), global)

	require.NoError(t, err)
	require.Equal(t, DefaultE2BSandboxTTL, got.E2BSandboxTTL)
	require.Equal(t, DefaultE2BHTTPTimeout, got.E2BHTTPTimeout)
	require.Equal(t, global.DefaultTimeout, got.DefaultTimeout,
		"the execution timeout is deployment policy and does carry over")
}

func TestResolveEffectiveConfigDisabled(t *testing.T) {
	tenantCfg := &types.TenantSandboxConfig{SandboxType: "disabled"}

	got, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.NoError(t, err)
	require.Equal(t, SandboxTypeDisabled, got.Type)
}

func TestResolveEffectiveConfigRejectsUnknownType(t *testing.T) {
	tenantCfg := &types.TenantSandboxConfig{SandboxType: "quantum"}

	_, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.Error(t, err)
}

func TestResolveEffectiveConfigRejectsUnsafeURL(t *testing.T) {
	tenantCfg := &types.TenantSandboxConfig{
		SandboxType: "e2b",
		E2B:         &types.E2BSandboxConfig{APIURL: "http://169.254.169.254"},
	}

	_, err := ResolveEffectiveConfig(tenantCfg, globalTestConfig())

	require.ErrorIs(t, err, ErrUnsafeOutboundURL)
}

func TestResolveEffectiveConfigRejectsDockerHostNetwork(t *testing.T) {
	_, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker: &types.DockerSandboxConfig{
			Image:       "weknora:test",
			NetworkMode: "host",
		},
	}, DefaultConfig())
	require.Error(t, err)
}

// A blank host must come out of config resolution already pointing at the
// daemon the Docker CLI would use, because that resolved value is what the
// connectivity check dials. Leaving it empty here is what made the check fail
// on a Colima or Docker Desktop host, where /var/run/docker.sock is absent.
func TestResolveEffectiveConfigDetectsLocalDockerHostWhenBlank(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix:///tmp/from-env.sock")

	effective, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker:      &types.DockerSandboxConfig{Image: "weknora:test"},
	}, DefaultConfig())

	require.NoError(t, err)
	require.Equal(t, "unix:///tmp/from-env.sock", effective.DockerHost)
}

func TestResolveEffectiveConfigMapsDockerNoneToDeniedEgress(t *testing.T) {
	effective, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker: &types.DockerSandboxConfig{
			Image:       "weknora:test",
			NetworkMode: "none",
		},
	}, DefaultConfig())
	require.NoError(t, err)
	require.True(t, effective.Network.DeniesEgressByDefault(),
		"docker network_mode=none is this backend's deny-all switch")
}

func TestResolveEffectiveConfigKeepsDockerBridgeEgressOpen(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix:///tmp/from-env.sock")
	effective, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker: &types.DockerSandboxConfig{
			Image:       "weknora:test",
			NetworkMode: "bridge",
		},
	}, DefaultConfig())
	require.NoError(t, err)
	require.False(t, effective.Network.DeniesEgressByDefault())
}

func TestResolveEffectiveConfigRejectsPlaintextDockerTCP(t *testing.T) {
	_, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType:           "docker",
		AllowPrivateEndpoints: true,
		Docker: &types.DockerSandboxConfig{
			Image: "weknora:test",
			Host:  "tcp://10.0.0.5:2376",
		},
	}, DefaultConfig())
	require.Error(t, err)
}

// The same bar has to apply to a host nobody typed. A blank field is filled in
// from DOCKER_HOST, and on a deployment pointed at a plaintext daemon that
// resolved value is what gets dialled — so validating only the stored string
// would accept a config that cannot work and say so only at the first sandbox.
func TestResolveEffectiveConfigRejectsResolvedPlaintextDockerTCP(t *testing.T) {
	t.Setenv("DOCKER_HOST", "tcp://10.0.0.5:2375")

	_, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType:           "docker",
		AllowPrivateEndpoints: true,
		Docker:                &types.DockerSandboxConfig{Image: "weknora:test"},
	}, DefaultConfig())
	require.Error(t, err)
}

// A resolved host is still only a daemon endpoint, so it answers to the same
// outbound policy an explicitly typed one does.
func TestResolveEffectiveConfigRejectsResolvedPrivateDockerHostWithoutOptIn(t *testing.T) {
	t.Setenv("DOCKER_HOST", "tcp://10.0.0.5:2376")

	_, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker: &types.DockerSandboxConfig{
			Image:       "weknora:test",
			TLSCertPath: "/etc/weknora/docker-certs",
		},
	}, DefaultConfig())
	require.ErrorIs(t, err, ErrUnsafeOutboundURL)
}

// A missing image is a field the admin can see and fix in the form, so it must
// still be reported ahead of anything about the daemon endpoint.
func TestResolveEffectiveConfigReportsMissingImageBeforeHostProblems(t *testing.T) {
	t.Setenv("DOCKER_HOST", "tcp://10.0.0.5:2375")

	_, err := ResolveEffectiveConfig(&types.TenantSandboxConfig{
		SandboxType: "docker",
	}, DefaultConfig())
	require.ErrorIs(t, err, ErrSandboxConfigIncomplete)
}

func TestEffectiveTemplateIDPerProvider(t *testing.T) {
	require.Equal(t, "e2b-tpl", EffectiveTemplateID(&Config{
		Type: SandboxTypeE2B, E2BTemplate: "e2b-tpl", DockerImage: "docker-img",
	}))
	require.Equal(t, "docker-img", EffectiveTemplateID(&Config{
		Type: SandboxTypeDocker, E2BTemplate: "e2b-tpl", DockerImage: "docker-img",
	}))
	require.Empty(t, EffectiveTemplateID(&Config{Type: SandboxTypeDisabled}))
	require.Empty(t, EffectiveTemplateID(nil))
}

func TestResolveEffectiveConfigUsesSkillSnapshotAsE2BTemplate(t *testing.T) {
	global := DefaultConfig()

	base := &types.TenantSandboxConfig{
		SandboxType: "e2b",
		E2B: &types.E2BSandboxConfig{
			APIURL: "https://203.0.113.20", SandboxDomain: "e2b.example.com",
			APIKey: "key-1", TemplateID: "tpl-base",
		},
	}
	fp := SkillImageFingerprint("e2b", "key-1", "https://203.0.113.20")
	cubeFp := SkillImageFingerprint("cube", "key-1", "https://203.0.113.20")

	t.Run("usable snapshot overrides the base template", func(t *testing.T) {
		cfg := *base
		cfg.SkillImage = &types.SkillImageConfig{SnapshotID: "snap-1", OwnerFingerprint: fp}

		eff, err := ResolveEffectiveConfig(&cfg, global)

		require.NoError(t, err)
		require.Equal(t, "snap-1", eff.E2BTemplate)
	})

	t.Run("fingerprint mismatch falls back to the base template", func(t *testing.T) {
		cfg := *base
		cfg.SkillImage = &types.SkillImageConfig{
			SnapshotID: "snap-1", OwnerFingerprint: cubeFp,
		}

		eff, err := ResolveEffectiveConfig(&cfg, global)

		require.NoError(t, err)
		require.Equal(t, "tpl-base", eff.E2BTemplate,
			"a snapshot whose fingerprint was computed for cube must not override e2b; the session must still boot")
	})

	t.Run("empty snapshot keeps the base template", func(t *testing.T) {
		cfg := *base
		cfg.SkillImage = &types.SkillImageConfig{OwnerFingerprint: fp}

		eff, err := ResolveEffectiveConfig(&cfg, global)

		require.NoError(t, err)
		require.Equal(t, "tpl-base", eff.E2BTemplate)
	})
}

func TestResolveEffectiveConfigUsesSkillSnapshotAsDockerImage(t *testing.T) {
	global := DefaultConfig()
	base := &types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker: &types.DockerSandboxConfig{
			Image: "weknora/sandbox:base",
			Host:  "unix:///var/run/docker.sock",
		},
	}
	fp := SkillImageFingerprint("docker", "", "unix:///var/run/docker.sock")

	t.Run("usable snapshot overrides the base image", func(t *testing.T) {
		cfg := *base
		cfg.Docker = &types.DockerSandboxConfig{
			Image: "weknora/sandbox:base", Host: "unix:///var/run/docker.sock",
		}
		cfg.SkillImage = &types.SkillImageConfig{
			SnapshotID: "weknora-skill/weknora-sk-cfg1-g1", OwnerFingerprint: fp,
		}

		eff, err := ResolveEffectiveConfig(&cfg, global)

		require.NoError(t, err)
		require.Equal(t, "weknora-skill/weknora-sk-cfg1-g1", eff.DockerImage)
	})

	t.Run("fingerprint mismatch falls back to the base image", func(t *testing.T) {
		cfg := *base
		cfg.Docker = &types.DockerSandboxConfig{
			Image: "weknora/sandbox:base", Host: "unix:///var/run/docker.sock",
		}
		cfg.SkillImage = &types.SkillImageConfig{
			SnapshotID: "weknora-skill/weknora-sk-cfg1-g1", OwnerFingerprint: "another-daemon",
		}

		eff, err := ResolveEffectiveConfig(&cfg, global)

		require.NoError(t, err)
		require.Equal(t, "weknora/sandbox:base", eff.DockerImage,
			"a snapshot from another daemon is invisible; the session must still boot")
	})

	// A blank host follows the environment, so binding the fingerprint to the
	// resolved daemon would make "switch docker context" indistinguishable
	// from "rotate the credentials": sessions would silently lose every
	// skill, installs would be refused, and the snapshot prune would skip the
	// config forever. The images are on the same disk either way.
	t.Run("blank host survives a change of local daemon", func(t *testing.T) {
		newConfig := func() *types.TenantSandboxConfig {
			return &types.TenantSandboxConfig{
				SandboxType: "docker",
				Docker:      &types.DockerSandboxConfig{Image: "weknora/sandbox:base"},
				SkillImage: &types.SkillImageConfig{
					SnapshotID: "weknora-skill/weknora-sk-cfg1-g1",
					OwnerFingerprint: SkillOwnerFingerprint(&types.TenantSandboxConfig{
						SandboxType: "docker",
						Docker:      &types.DockerSandboxConfig{Image: "weknora/sandbox:base"},
					}),
				},
			}
		}

		for _, host := range []string{
			"unix:///tmp/colima.sock",
			"unix:///tmp/orbstack.sock",
		} {
			t.Setenv("DOCKER_HOST", host)
			cfg := newConfig()

			eff, err := ResolveEffectiveConfig(cfg, global)

			require.NoError(t, err)
			require.Equal(t, "weknora-skill/weknora-sk-cfg1-g1", eff.DockerImage,
				"DOCKER_HOST=%s must not retire the config's skill image", host)
			require.True(t, SkillImageActive(cfg))
		}
	})

	// An explicit host is still part of the identity: only an admin editing
	// the config can change it, and it may well be another machine.
	t.Run("an explicit host change retires the image", func(t *testing.T) {
		cfg := *base
		cfg.Docker = &types.DockerSandboxConfig{
			Image: "weknora/sandbox:base", Host: "tcp://198.51.100.10:2376",
			TLSCertPath: "/certs",
		}
		cfg.SkillImage = &types.SkillImageConfig{
			SnapshotID: "weknora-skill/weknora-sk-cfg1-g1", OwnerFingerprint: fp,
		}

		eff, err := ResolveEffectiveConfig(&cfg, global)

		require.NoError(t, err)
		require.Equal(t, "weknora/sandbox:base", eff.DockerImage)
	})
}

func TestSkillOwnerFingerprintForDocker(t *testing.T) {
	require.Empty(t, SkillOwnerFingerprint(&types.TenantSandboxConfig{SandboxType: "docker"}),
		"a docker type with no daemon block cannot own a snapshot")
	require.Empty(t, SkillOwnerFingerprint(&types.TenantSandboxConfig{SandboxType: "disabled"}))

	got := SkillOwnerFingerprint(&types.TenantSandboxConfig{
		SandboxType: "docker",
		Docker: &types.DockerSandboxConfig{
			Image: "img", Host: "unix:///var/run/docker.sock", TLSCertPath: "/certs",
		},
	})
	require.Equal(t, SkillImageFingerprint("docker", "/certs", "unix:///var/run/docker.sock"), got)
}

// SkillImageActive is what the agent side asks before telling a model about an
// installed skill, so it must agree with the template ResolveEffectiveConfig
// actually boots. Any disagreement means either skills that are announced and
// cannot run, or skills that are in the image and hidden.
func TestSkillImageFingerprintIsStableAndDiscriminating(t *testing.T) {
	a := SkillImageFingerprint("cube", "key-1", "https://a.example.com")
	require.Equal(t, a, SkillImageFingerprint("cube", "key-1", "https://a.example.com"))
	require.NotEqual(t, a, SkillImageFingerprint("cube", "key-2", "https://a.example.com"))
	require.NotEqual(t, a, SkillImageFingerprint("e2b", "key-1", "https://a.example.com"))
}

func TestParseSandboxTypeAcceptsHost(t *testing.T) {
	got, err := ParseSandboxType("host")
	require.NoError(t, err)
	require.Equal(t, SandboxTypeHost, got)
}

// host carries no endpoint or credential, so a bare config is complete.
func TestHostConfigNeedsNoProviderFields(t *testing.T) {
	require.Empty(t, MissingRequiredFields(&Config{Type: SandboxTypeHost}))
}
