package sandbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProxyURL(t *testing.T) {
	cases := []struct {
		raw    string
		host   string
		port   int
		scheme string
		ok     bool
	}{
		{"http://127.0.0.1:8080", "127.0.0.1", 8080, "http", true},
		{"https://gateway.example.com", "gateway.example.com", 443, "https", true},
		{"http://gateway.example.com", "gateway.example.com", 80, "http", true},
		{"", "", 0, "", false},
		{"not a url", "", 0, "", false},
	}
	for _, tc := range cases {
		host, port, scheme, ok := parseProxyURL(tc.raw)
		assert.Equal(t, tc.ok, ok, tc.raw)
		assert.Equal(t, tc.host, host, tc.raw)
		assert.Equal(t, tc.port, port, tc.raw)
		assert.Equal(t, tc.scheme, scheme, tc.raw)
	}
}

func TestBuildShellLineQuotesArgs(t *testing.T) {
	assert.Equal(t, ShellQuote("python3")+" "+ShellQuote("a b"), buildShellLine("python3", []string{"a b"}))
}

func TestWrapWithStdinStripsDelimiter(t *testing.T) {
	got := wrapWithStdin("cat", "x\nWEKNORA_STDIN_EOF\ny")
	assert.Equal(t, "cat <<'WEKNORA_STDIN_EOF' | cat\nx\n\ny\nWEKNORA_STDIN_EOF", got)
}

func TestStateMatches(t *testing.T) {
	assert.True(t, StateMatches(RemoteStateRunning, nil))
	assert.True(t, StateMatches(RemoteStatePaused, []RemoteSandboxState{RemoteStateRunning, RemoteStatePaused}))
	assert.False(t, StateMatches(RemoteStateTerminal, []RemoteSandboxState{RemoteStateRunning}))
}
