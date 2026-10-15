package container

import (
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/datasource"
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

func TestConnectorRegistryIncludesSeafile(t *testing.T) {
	registry, err := initConnectorRegistry()
	if err != nil {
		t.Fatalf("initConnectorRegistry() error = %v", err)
	}
	connector, err := registry.Get(types.ConnectorTypeSeafile)
	if err != nil {
		t.Fatalf("Seafile connector is not registered: %v", err)
	}
	if connector.Type() != types.ConnectorTypeSeafile {
		t.Fatalf("connector.Type() = %q", connector.Type())
	}
	if _, ok := connector.(datasource.FullStreamingConnector); !ok {
		t.Fatal("Seafile connector must implement datasource.FullStreamingConnector")
	}
}
