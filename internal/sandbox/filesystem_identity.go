package sandbox

import (
	"context"
	"io"
	"strings"
)

// Only the server's skill installer may select the maintenance identity.
// Ordinary filesystem operations use the same account as shell/script exec.
type maintenanceFilesystemKey struct{}

func withMaintenanceFilesystem(ctx context.Context) context.Context {
	return context.WithValue(ctx, maintenanceFilesystemKey{}, true)
}

func remoteFileUser(ctx context.Context) string {
	if maintenance, _ := ctx.Value(maintenanceFilesystemKey{}).(bool); maintenance {
		return "root"
	}
	return DefaultSandboxExecUser
}

type cancelOnCloseBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelOnCloseBody) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}

func rpcUsesCallerDeadline(path string) bool {
	if strings.HasPrefix(path, "/process.Process/") {
		return true
	}
	if isFilesystemContentRPC(path) {
		return true
	}
	return false
}

func isFilesystemContentRPC(path string) bool {
	if path == envdFilesRoute {
		return true
	}
	name, ok := strings.CutPrefix(path, "/filesystem.Filesystem/")
	if !ok {
		return false
	}
	return strings.HasPrefix(name, "Read") || strings.HasPrefix(name, "Write")
}
