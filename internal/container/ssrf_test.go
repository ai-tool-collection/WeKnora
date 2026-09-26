package container

import (
	"os"
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/types"
	"github.com/ai-tool-collection/WeKnora/internal/utils"
)

func TestMain(m *testing.M) {
	// Engine wiring tests intentionally use loopback httptest servers.
	utils.SetSSRFWhitelistFromRaw("127.0.0.1,::1,localhost")
	code := m.Run()
	utils.SetSSRFWhitelistFromRaw("")
	os.Exit(code)
}

func TestValidateRuntimeVectorStoreAddressesAllowsNonNetworkEngines(t *testing.T) {
	for _, engineType := range []types.RetrieverEngineType{
		types.PostgresRetrieverEngineType,
		types.SQLiteRetrieverEngineType,
	} {
		if err := validateRuntimeVectorStoreAddresses(types.VectorStore{EngineType: engineType}); err != nil {
			t.Fatalf("unexpected %s validation error: %v", engineType, err)
		}
	}
}
