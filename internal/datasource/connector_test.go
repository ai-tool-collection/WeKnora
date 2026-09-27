package datasource

import (
	"testing"
)

func TestConnectorMetadataDoesNotAdvertiseWebhook(t *testing.T) {
	for connectorType, meta := range ConnectorMetadataRegistry {
		for _, capability := range meta.Capabilities {
			if capability == "webhook" {
				t.Fatalf("%s connector should not advertise webhook until webhook sync is implemented", connectorType)
			}
		}
	}
}

// Removed China-hosted connectors must not come back through an upstream merge.
func TestConnectorMetadataExcludesRemovedConnectors(t *testing.T) {
	for _, removed := range []string{
		"feishu", "lark", "feishu_drive", "lark_drive", "yuque", "dingtalk", "ima",
	} {
		if _, ok := ConnectorMetadataRegistry[removed]; ok {
			t.Fatalf("removed connector %q is registered", removed)
		}
	}
}
