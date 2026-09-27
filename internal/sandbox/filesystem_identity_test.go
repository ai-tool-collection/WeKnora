package sandbox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEnvdMaintenanceFilesystemIdentity(t *testing.T) {
	recorder := &recordingRoundTripper{}
	transport := NewEnvdCompatTransport(recorder, DefaultSandboxExecUser)
	req := httptest.NewRequest(http.MethodGet, "https://sandbox.example/files", nil).
		WithContext(withMaintenanceFilesystem(context.Background()))
	_, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.Equal(t, basicAuthorizationFor("root"), recorder.request.Header.Get("Authorization"))
}

func TestWorkspaceBootstrapDoesNotRemoveOrMoveFiles(t *testing.T) {
	root := t.TempDir()
	blocked := filepath.Join(root, "existing-file")
	require.NoError(t, os.WriteFile(blocked, []byte("keep this"), 0600))
	cmd := exec.Command("/bin/sh", "-c", workspaceBootstrapCommand(blocked))
	require.Error(t, cmd.Run())
	content, err := os.ReadFile(blocked)
	require.NoError(t, err)
	require.Equal(t, "keep this", string(content))
	link := filepath.Join(root, "link")
	require.NoError(t, os.Symlink(root, link))
	require.Error(t, exec.Command("/bin/sh", "-c", workspaceBootstrapCommand(link)).Run())
	_, err = os.Lstat(link)
	require.NoError(t, err)
	valid := filepath.Join(root, "space ' quote", "nested")
	require.NoError(t, exec.Command("/bin/sh", "-c", workspaceBootstrapCommand(valid)).Run())
	info, err := os.Stat(valid)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestFileTransferSkipsShortHTTPTimeout(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assertBudget := func(t *testing.T, transport http.RoundTripper, route string, short bool) {
		t.Helper()
		recorder := &recordingRoundTripper{}
		typed, ok := transport.(*e2bRPCTimeoutTransport)
		require.True(t, ok, "unexpected transport %T", transport)
		typed.next = recorder
		req := httptest.NewRequest(http.MethodGet, "https://sandbox.example"+route, nil).WithContext(parent)
		_, err := transport.RoundTrip(req)
		require.NoError(t, err)
		deadline, ok := recorder.request.Context().Deadline()
		require.True(t, ok)
		remaining := time.Until(deadline)
		if short {
			require.Less(t, remaining, 200*time.Millisecond)
			return
		}
		require.Greater(t, remaining, time.Second)
	}

	e2b := &e2bRPCTimeoutTransport{timeout: 50 * time.Millisecond}
	assertBudget(t, e2b, "/files", false)
	assertBudget(t, e2b, "/filesystem.Filesystem/Read", false)
	assertBudget(t, e2b, "/filesystem.Filesystem/MakeDir", true)
	assertBudget(t, e2b, "/process.Process/List", true)
	assertBudget(t, e2b, "/process.Process/Start", false)
}
