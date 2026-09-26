package parity

import (
	"testing"

	modelruntime "github.com/ai-tool-collection/WeKnora/internal/models/runtime"
	"github.com/stretchr/testify/require"
)

// The registered model providers must match the catalog shipped by this fork.
func TestRegisteredProviders(t *testing.T) {
	allowed := map[string]bool{
		"anthropic": true, "azure_openai": true, "gemini": true,
		"generic": true, "gpustack": true, "jina": true,
		"litellm": true, "novita": true, "nvidia": true,
		"openai": true, "openrouter": true, "requesty": true,
	}
	for _, provider := range modelruntime.List() {
		require.True(t, allowed[provider.ID], "unexpected provider %s", provider.ID)
		delete(allowed, provider.ID)
	}
	require.Empty(t, allowed, "registered providers are missing")
}
