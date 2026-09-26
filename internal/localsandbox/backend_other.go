//go:build !darwin && !windows

package localsandbox

import (
	"context"
	"fmt"

	"github.com/ai-tool-collection/WeKnora/internal/localsandbox/core"
	"github.com/ai-tool-collection/WeKnora/internal/logger"
)

// NewBackend has no implementation outside macOS and Windows.
func NewBackend() (core.Backend, error) {
	err := fmt.Errorf("%w: only macOS is supported today", core.ErrUnsupportedPlatform)
	logger.Errorf(context.Background(), "[LocalSandbox] %v", err)
	return nil, err
}
