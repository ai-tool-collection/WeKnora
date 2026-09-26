package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The zero value must mean "egress open, inbound closed", which is what lets a
// config with no network block behave identically to one saved with defaults.
func TestSandboxNetworkPolicyZeroValueIsTheDefault(t *testing.T) {
	var p SandboxNetworkPolicy

	require.False(t, p.DenyEgressByDefault, "egress must default to allowed")
	require.False(t, p.AllowPublicInbound, "inbound must default to closed")

	encoded, err := json.Marshal(&p)
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(encoded),
		"a default policy must serialize empty so it is indistinguishable from nil")
}

func TestSandboxNetworkPolicyCloneWithSecretsNil(t *testing.T) {
	var p *SandboxNetworkPolicy
	require.Nil(t, p.CloneWithSecrets(func(s string) string { return s }))
}

func firstKey(m map[string]string) string {
	for k := range m {
		return k
	}
	return ""
}

func TestValidateSandboxNetworkPolicyAcceptsDefaults(t *testing.T) {
	require.NoError(t, ValidateSandboxNetworkPolicy(nil))
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{SandboxType: "cube"}))
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network:     &SandboxNetworkPolicy{},
	}))
}

func TestValidateSandboxNetworkPolicyDomainAllowNeedsDenyAll(t *testing.T) {
	// A domain allow-list without a deny-all is misleading rather than strict:
	// destinations the sandbox never resolved through DNS stay reachable.
	err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network:     &SandboxNetworkPolicy{AllowOut: []string{"api.example.com"}},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "0.0.0.0/0")

	// Either of the two ways to express deny-all is enough.
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			AllowOut:            []string{"api.example.com"},
		},
	}))
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network: &SandboxNetworkPolicy{
			AllowOut: []string{"api.example.com"},
			DenyOut:  []string{"0.0.0.0/0"},
		},
	}))
	// IP-only allow lists never need the fallback.
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network:     &SandboxNetworkPolicy{AllowOut: []string{"1.1.1.1", "203.0.113.0/24"}},
	}))
}

// DenyOutCoversAllIPv4 exists so every consumer of the policy agrees with the
// validator about what "deny all" looks like. It must therefore use the same
// normalisation the validator does, not a string compare: net.ParseCIDR
// collapses any /0 onto 0.0.0.0/0, and the validator trims whitespace.
func TestDenyOutCoversAllIPv4MatchesValidatorNormalisation(t *testing.T) {
	require.True(t, DenyOutCoversAllIPv4([]string{"0.0.0.0/0"}))
	require.True(t, DenyOutCoversAllIPv4([]string{"  0.0.0.0/0  "}))
	require.True(t, DenyOutCoversAllIPv4([]string{"1.2.3.4/0"}))
	require.True(t, DenyOutCoversAllIPv4([]string{"10.0.0.0/8", "0.0.0.0/0"}))

	require.False(t, DenyOutCoversAllIPv4(nil))
	require.False(t, DenyOutCoversAllIPv4([]string{}))
	require.False(t, DenyOutCoversAllIPv4([]string{"10.0.0.0/8"}))
	require.False(t, DenyOutCoversAllIPv4([]string{"0.0.0.0/32"}))
	require.False(t, DenyOutCoversAllIPv4([]string{"0.0.0.0"}))
	// Unparseable entries are the validator's problem, not this predicate's.
	require.False(t, DenyOutCoversAllIPv4([]string{"not a cidr"}))

	require.Equal(t, []string{DenyAllIPv4}, CanonicalizeDenyOut([]string{"1.2.3.4/0"}))
	require.Equal(t, []string{DenyAllIPv4}, CanonicalizeDenyOut([]string{"  0.0.0.0/0  "}))
	require.Equal(t, []string{"10.0.0.0/8", DenyAllIPv4}, CanonicalizeDenyOut([]string{"10.0.0.0/8", "1.2.3.4/0"}))
	require.Nil(t, CanonicalizeDenyOut(nil))

	// The two spellings the validator accepts must agree with the predicate.
	for _, policy := range []*SandboxNetworkPolicy{
		{AllowOut: []string{"api.example.com"}, DenyOut: []string{"0.0.0.0/0"}},
		{AllowOut: []string{"api.example.com"}, DenyOut: []string{"1.2.3.4/0"}},
	} {
		require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
			SandboxType: "cube",
			Network:     policy,
		}))
		require.True(t, DenyOutCoversAllIPv4(policy.DenyOut))
	}
}

func TestValidateSandboxNetworkPolicyRejectsDomainInDenyOut(t *testing.T) {
	err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network:     &SandboxNetworkPolicy{DenyOut: []string{"api.example.com"}},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "deny_out")
}

func TestValidateSandboxNetworkPolicyRejectsBadTargets(t *testing.T) {
	for _, target := range []string{
		"*", ".example.com", "*.*.example.com",
		"api.*.example.com", "example.com:443", "999.999.999.999", "*.999.999", "",
	} {
		err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
			SandboxType: "cube",
			Network: &SandboxNetworkPolicy{
				DenyEgressByDefault: true,
				AllowOut:            []string{target},
			},
		})
		require.Error(t, err, "target %q must be rejected", target)
	}
}

func TestValidateSandboxNetworkPolicyE2BHostRuleNeedsAllowOut(t *testing.T) {
	// A transform rule grants no egress on its own.
	err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "e2b",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			E2BHostRules: []E2BHostRule{{
				Host:    "api.example.com",
				Headers: map[string]string{"X-Key": "v"},
			}},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "allow_out")

	// A wildcard allow covers its subdomains.
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "e2b",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			AllowOut:            []string{"*.example.com"},
			E2BHostRules: []E2BHostRule{{
				Host:    "api.example.com",
				Headers: map[string]string{"X-Key": "v"},
			}},
		},
	}))

	for _, host := range []string{"example.com", "evilexample.com"} {
		err = ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
			SandboxType: "e2b",
			Network: &SandboxNetworkPolicy{
				DenyEgressByDefault: true,
				AllowOut:            []string{"*.example.com"},
				E2BHostRules:        []E2BHostRule{{Host: host}},
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "allow_out")
	}
}

func TestValidateSandboxNetworkPolicyE2BLimits(t *testing.T) {
	headers := make(map[string]string, 21)
	for i := 0; i < 21; i++ {
		headers["X-Header-"+string(rune('a'+i))] = "v"
	}
	err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "e2b",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			AllowOut:            []string{"api.example.com"},
			E2BHostRules:        []E2BHostRule{{Host: "api.example.com", Headers: headers}},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "最多 20 个 header")

	domains := make([]string, e2bMaxRuleDomains+1)
	rules := make([]E2BHostRule, e2bMaxRuleDomains+1)
	for i := range domains {
		domains[i] = string(rune('a'+i)) + ".example.com"
		rules[i] = E2BHostRule{Host: domains[i]}
	}
	err = ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "e2b",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			AllowOut:            domains,
			E2BHostRules:        rules,
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "最多 10 个 host 规则域名")

	longHost := strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + ".com"
	err = ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "e2b",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			E2BHostRules:        []E2BHostRule{{Host: longHost}},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "超过 128 字符")

	for _, tc := range []struct {
		name, value, want string
	}{
		{strings.Repeat("h", e2bMaxHeaderNameLength+1), "v", "header 名超过 64 字符"},
		{"X-Key", strings.Repeat("v", e2bMaxHeaderValueLen+1), "header 值超过 2048 字符"},
	} {
		err = ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
			SandboxType: "e2b",
			Network: &SandboxNetworkPolicy{
				DenyEgressByDefault: true,
				AllowOut:            []string{"api.example.com"},
				E2BHostRules: []E2BHostRule{{
					Host:    "api.example.com",
					Headers: map[string]string{tc.name: tc.value},
				}},
			},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), tc.want)
	}
}

func TestValidateSandboxNetworkPolicyDockerRejectsFineGrained(t *testing.T) {
	err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "docker",
		Network:     &SandboxNetworkPolicy{AllowOut: []string{"1.1.1.1"}},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "docker")

	// The overall switches remain meaningful for Docker.
	require.NoError(t, ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "docker",
		Network:     &SandboxNetworkPolicy{DenyEgressByDefault: true},
	}))
}

func TestValidateSandboxNetworkPolicyRejectsDuplicateTargets(t *testing.T) {
	// Cube normalises map keys, so these two are one entry; accepting both
	// would let an admin believe they configured two.
	err := ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network: &SandboxNetworkPolicy{
			AllowOut: []string{"198.51.100.1", "198.51.100.1/32"},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "198.51.100.1")

	err = ValidateSandboxNetworkPolicy(&TenantSandboxConfig{
		SandboxType: "cube",
		Network: &SandboxNetworkPolicy{
			DenyEgressByDefault: true,
			AllowOut:            []string{"API.Example.com.", "api.example.com"},
		},
	})
	require.Error(t, err)
}
