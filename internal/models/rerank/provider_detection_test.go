package rerank

import (
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/models/providers"
	modelruntime "github.com/ai-tool-collection/WeKnora/internal/models/runtime"
)

// TestDetectByURLSeesVendorCatalog guards the blank import of
// runtime composition.
//
// modelruntime.DetectByURL returns "generic" for every URL while the catalog is
// empty. Without the vendor packages linked in, a stored rerank row that
// carries no provider id may resolve to the generic client instead of its
// provider-specific protocol.
func TestDetectByURLSeesVendorCatalog(t *testing.T) {
	cases := map[string]string{
		"https://api.jina.ai/v1":                               "jina",
		"https://some-self-hosted-gateway.example.internal/v1": providers.GenericID,
	}
	for baseURL, want := range cases {
		if got := modelruntime.DetectByURL(baseURL); got != want {
			t.Errorf("DetectByURL(%q) = %q, want %q", baseURL, got, want)
		}
	}
}
