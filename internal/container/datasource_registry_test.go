package container

import (
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/types"
)

func TestConnectorRegistryIncludesRetainedConnectors(t *testing.T) {
	registry, err := initConnectorRegistry()
	if err != nil {
		t.Fatalf("initConnectorRegistry() error = %v", err)
	}
	for _, connectorType := range []string{
		types.ConnectorTypeNotion, types.ConnectorTypeConfluence, types.ConnectorTypeRSS, types.ConnectorTypeGitLab,
	} {
		connector, err := registry.Get(connectorType)
		if err != nil {
			t.Fatalf("%s connector is not registered: %v", connectorType, err)
		}
		if connector.Type() != connectorType {
			t.Fatalf("connector.Type() = %q, want %q", connector.Type(), connectorType)
		}
	}
	// Removed China-hosted connectors must stay out of the registry.
	for _, removed := range []string{"feishu", "lark", "feishu_drive", "lark_drive", "yuque", "dingtalk", "ima"} {
		if _, err := registry.Get(removed); err == nil {
			t.Fatalf("removed connector %q is registered", removed)
		}
	}
}
