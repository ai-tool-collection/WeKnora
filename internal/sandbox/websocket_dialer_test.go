package sandbox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// desktopEchoServer is a websockify stand-in: it records the handshake
// request and accepts the upgrade.
func desktopEchoServer(t *testing.T, seen *http.Request) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{
		CheckOrigin:  func(*http.Request) bool { return true },
		Subprotocols: []string{"binary"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = *r.Clone(r.Context())
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		_, _, _ = conn.NextReader()
	}))
}

func TestWebsocketDialerE2BEmptyDomainUsesSDKDefault(t *testing.T) {
	// E2B Cloud configs may omit sandbox_domain: go-e2b fills e2b.app for
	// envd, and the desktop Host header has to match that same default.
	var seen http.Request
	srv := desktopEchoServer(t, &seen)
	defer srv.Close()

	pool := NewSandboxGatewayTransportPoolWithPolicy(nil, OutboundURLPolicy{AllowPrivate: true})
	pool.InboundTokens().Put("sbx-1", "tok")

	cfg := &Config{
		Type:                  SandboxTypeE2B,
		E2BProxyURL:           srv.URL,
		AllowPrivateEndpoints: true,
	}
	conn, _, err := pool.WebsocketDialerFor(cfg).
		Dial(context.Background(), "sbx-1", 6080, "/websockify", nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	require.Equal(t, "6080-sbx-1.e2b.app", seen.Host)
}

func TestWebsocketDialerE2BConfiguredDomainWinsOverDefault(t *testing.T) {
	var seen http.Request
	srv := desktopEchoServer(t, &seen)
	defer srv.Close()

	pool := NewSandboxGatewayTransportPoolWithPolicy(nil, OutboundURLPolicy{AllowPrivate: true})
	pool.InboundTokens().Put("sbx-1", "tok")

	cfg := &Config{
		Type:                  SandboxTypeE2B,
		E2BProxyURL:           srv.URL,
		E2BSandboxDomain:      "sandbox.internal",
		AllowPrivateEndpoints: true,
	}
	conn, _, err := pool.WebsocketDialerFor(cfg).
		Dial(context.Background(), "sbx-1", 6080, "/websockify", nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	require.Equal(t, "6080-sbx-1.sandbox.internal", seen.Host)
}

func TestE2BDialDesktopWithoutPoolIsUnsupported(t *testing.T) {
	// NewE2BRemoteClientWithTransport has no pool, so it has no dialer. It
	// must say "unsupported" rather than nil-panic.
	client, err := NewE2BRemoteClientWithTransport(&Config{
		Type: SandboxTypeE2B, E2BAPIKey: "k", E2BTemplate: "tpl",
	}, nil)
	require.NoError(t, err)

	_, derr := client.DialDesktop(context.Background(), nil, RemoteDesktopOptions{})
	require.Error(t, derr)
	var remoteErr *RemoteError
	require.ErrorAs(t, derr, &remoteErr)
	require.Equal(t, RemoteErrorKindUnsupported, remoteErr.Kind)
	require.Equal(t, "DialDesktop", remoteErr.Op)
}
