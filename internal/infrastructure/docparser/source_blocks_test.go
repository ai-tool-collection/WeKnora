package docparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ai-tool-collection/WeKnora/docreader/proto"
	"github.com/ai-tool-collection/WeKnora/internal/types"
)

func TestSourceBlocksFromProto(t *testing.T) {
	blocks := sourceBlocksFromProto([]*proto.SourceBlock{
		{Start: 0, End: 5, LocatorJson: `{"type":"pdf","page":2,"bbox":[0.1,0.2,0.3,0.4]}`},
		{Start: 5, End: 5, LocatorJson: `{"type":"pdf","page":2}`},
		{Start: 5, End: 9, LocatorJson: `not json`},
	})
	require.Len(t, blocks, 1)
	assert.Equal(t, types.SourceLocator{Type: "pdf", Page: 2, BBox: []float64{0.1, 0.2, 0.3, 0.4}}, blocks[0].Locator)
}
