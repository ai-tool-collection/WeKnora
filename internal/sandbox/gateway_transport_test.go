package sandbox

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// countingRoundTripper records how many requests rode the control transport.
type countingRoundTripper struct {
	next  http.RoundTripper
	mu    sync.Mutex
	hosts []string
}

func (c *countingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	c.hosts = append(c.hosts, req.URL.Host)
	c.mu.Unlock()
	return c.next.RoundTrip(req)
}

func (c *countingRoundTripper) seen() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.hosts...)
}

// The data plane addresses sandboxes as "49983-{id}.{domain}" but must dial
// the configured proxy. Sharing one transport across both planes drops that
// rewrite, which is exactly the regression this guards: the proxy has to see
// the request, and it has to see the sandbox authority in the Host header.
func TestSandboxGatewayTransportPoolReusesTransportPerProxyEndpoint(t *testing.T) {
	pool := NewSandboxGatewayTransportPoolWithPolicy(
		NewGuardedTransportWithPolicy(OutboundURLPolicy{AllowPrivate: true}),
		OutboundURLPolicy{AllowPrivate: true},
	)

	first := pool.RoundTripperFor(&Config{
		Type:             SandboxTypeE2B,
		E2BProxyURL:      "http://127.0.0.1:8080",
		E2BSandboxDomain: "sandbox.example",
	}).(*gatewaySplitTransport)
	second := pool.RoundTripperFor(&Config{
		Type:             SandboxTypeE2B,
		E2BProxyURL:      "http://127.0.0.1:8080",
		E2BSandboxDomain: "sandbox.example",
	}).(*gatewaySplitTransport)
	other := pool.RoundTripperFor(&Config{
		Type:             SandboxTypeE2B,
		E2BProxyURL:      "http://127.0.0.1:9090",
		E2BSandboxDomain: "sandbox.example",
	}).(*gatewaySplitTransport)

	require.Same(t, first.data, second.data)
	require.NotSame(t, first.data, other.data)
	require.Same(t, first.control, other.control)
}

// Without a usable proxy URL the SDK dials the sandbox authority directly, so
// the split transport must not invent a data plane.
func TestSandboxGatewayTransportPoolWithoutProxyKeepsEverythingOnControl(t *testing.T) {
	pool := NewSandboxGatewayTransportPool(NewGuardedTransport())

	split := pool.RoundTripperFor(&Config{Type: SandboxTypeE2B, E2BSandboxDomain: "sandbox.example"}).(*gatewaySplitTransport)

	// A nil data transport is what sends sandbox authorities back to control.
	require.Nil(t, split.data)
}

// A self-hosted E2B-compatible control plane fronts every sandbox with one
// gateway, so the E2B provider must read its own gateway fields - and a plain
// HTTP gateway has to survive the SDK pinning the data-plane scheme to https.
func TestSandboxGatewayTransportPoolRoutesE2BDataPlane(t *testing.T) {
	pool := NewSandboxGatewayTransportPoolWithPolicy(
		NewGuardedTransportWithPolicy(OutboundURLPolicy{AllowPrivate: true}),
		OutboundURLPolicy{AllowPrivate: true},
	)

	split := pool.RoundTripperFor(&Config{
		Type:             SandboxTypeE2B,
		E2BProxyURL:      "http://127.0.0.1:18080",
		E2BSandboxDomain: "localhost",
	}).(*gatewaySplitTransport)

	require.NotNil(t, split.data)
	require.Equal(t, "http", split.dataScheme)
	require.True(t, split.isDataPlane("49983-sbx.localhost"))

	// Asserting on the address the transport dials pins the configured gateway
	// directly rather than relying on sandbox domain matching.
	require.Same(t, split.data, pool.dataTransport("127.0.0.1:18080"))
	require.NotSame(t, split.data, pool.dataTransport("127.0.0.1:9999"))
}

// The configured sandbox domain is not a reliable way to recognise data-plane
// traffic. sandbox_domain is optional for E2B (MissingRequiredFields does not
// demand it) and, more fundamentally, go-e2b's envdBaseURL prefers the domain
// the control plane reported over the client-wide one — so the authority the
// SDK actually dials can differ from anything in the config. Matching on it
// meant a configured gateway was silently never used.
func TestGatewaySplitTransportRoutesDataPlaneWithoutConfiguredDomain(t *testing.T) {
	pool := NewSandboxGatewayTransportPoolWithPolicy(
		NewGuardedTransportWithPolicy(OutboundURLPolicy{AllowPrivate: true}),
		OutboundURLPolicy{AllowPrivate: true},
	)

	split := pool.RoundTripperFor(&Config{
		Type:        SandboxTypeE2B,
		E2BProxyURL: "http://127.0.0.1:18080",
	}).(*gatewaySplitTransport)

	require.NotNil(t, split.data)
	require.True(t, split.isDataPlane("49983-sbx-1.e2b.app"),
		"an empty sandbox_domain must not disable the gateway")
}

// Same root cause, reachable even with the domain filled in: the control plane
// may report a domain the config never mentioned.
func TestGatewaySplitTransportRoutesDataPlaneOnServerReportedDomain(t *testing.T) {
	split := &gatewaySplitTransport{
		control: NewGuardedTransport(),
		data:    NewGuardedTransport(),
	}

	require.True(t, split.isDataPlane("49983-sbx-1.reported-by-server.example"))
}

// E2B's own defaults are api.e2b.app plus sandbox domain e2b.app, so suffix
// matching classified every control-plane call as data-plane traffic and dialled
// it at the gateway — sending X-API-Key to the wrong endpoint, and downgrading
// it to plaintext when the gateway is http://.
func TestGatewaySplitTransportKeepsControlHostOnControlTransport(t *testing.T) {
	pool := NewSandboxGatewayTransportPoolWithPolicy(
		NewGuardedTransportWithPolicy(OutboundURLPolicy{AllowPrivate: true}),
		OutboundURLPolicy{AllowPrivate: true},
	)

	split := pool.RoundTripperFor(&Config{
		Type:             SandboxTypeE2B,
		E2BAPIURL:        "https://api.mydomain.com",
		E2BSandboxDomain: "mydomain.com",
		E2BProxyURL:      "http://127.0.0.1:18080",
	}).(*gatewaySplitTransport)

	require.False(t, split.isDataPlane("api.mydomain.com"))
	require.True(t, split.isDataPlane("49983-sbx-1.mydomain.com"))
}

// An API host that happens to wear the data-plane shape is why the control host
// is excluded explicitly rather than left to the shape test.
func TestGatewaySplitTransportExcludesControlHostShapedLikeSandbox(t *testing.T) {
	pool := NewSandboxGatewayTransportPoolWithPolicy(
		NewGuardedTransportWithPolicy(OutboundURLPolicy{AllowPrivate: true}),
		OutboundURLPolicy{AllowPrivate: true},
	)

	split := pool.RoundTripperFor(&Config{
		Type:             SandboxTypeE2B,
		E2BAPIURL:        "https://8080-api.mydomain.com",
		E2BSandboxDomain: "mydomain.com",
		E2BProxyURL:      "http://127.0.0.1:18080",
	}).(*gatewaySplitTransport)

	require.False(t, split.isDataPlane("8080-api.mydomain.com"))
}

func TestGatewaySplitTransportAppliesGatewayScheme(t *testing.T) {
	recorder := &recordingRoundTripper{}
	split := &gatewaySplitTransport{
		control:    &recordingRoundTripper{},
		data:       recorder,
		dataScheme: "http",
	}

	request := httptest.NewRequest(http.MethodGet, "https://49983-sbx.localhost/files", nil)
	_, err := split.RoundTrip(request)
	require.NoError(t, err)

	require.Equal(t, "http", recorder.request.URL.Scheme)
	// The sandbox authority is what the gateway routes on; it must survive.
	require.Equal(t, "49983-sbx.localhost", recorder.request.URL.Host)
	require.Equal(t, "https", request.URL.Scheme, "the caller's request must not be mutated")
}
