//go:build windows

package localsandbox

import (
	"github.com/ai-tool-collection/WeKnora/internal/localsandbox/core"
	"github.com/ai-tool-collection/WeKnora/internal/localsandbox/winhost"
)

func NewBackend() (core.Backend, error) { return winhost.New() }
