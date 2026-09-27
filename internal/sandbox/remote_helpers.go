package sandbox

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

// parseProxyURL turns "http://127.0.0.1:80" into ("127.0.0.1", 80, "http").
// A missing port defaults to 80/443 depending on the scheme; an unparseable
// URL returns ok=false so callers can fall back to the SDK's defaults.
func parseProxyURL(raw string) (host string, port int, scheme string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return "", 0, "", false
	}
	scheme = strings.ToLower(parsed.Scheme)
	if scheme == "" {
		scheme = "http"
	}
	h, p, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		h = parsed.Host
		if scheme == "https" {
			p = "443"
		} else {
			p = "80"
		}
	}
	portInt, err := strconv.Atoi(p)
	if err != nil || portInt <= 0 {
		return "", 0, "", false
	}
	return h, portInt, scheme, true
}

// buildShellLine turns argv into a single shell-safe command line. It matches
// the semantics of the old hand-rolled path, which relied on envd's bash to
// resolve `python3` (or similar) against $PATH inside the sandbox image.
func buildShellLine(cmd string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, ShellQuote(cmd))
	for _, a := range args {
		parts = append(parts, ShellQuote(a))
	}
	return strings.Join(parts, " ")
}

// wrapWithStdin funnels a caller-supplied stdin payload into the child
// process by prepending a heredoc. This keeps the SDK's Commands.Run contract
// (which does not take an explicit stdin argument) usable in the rare case
// callers actually need to pipe data.
func wrapWithStdin(line, stdin string) string {
	// Use a heredoc delimiter unlikely to appear in caller data.
	const delim = "WEKNORA_STDIN_EOF"
	// Escape lines containing the delimiter defensively.
	safe := strings.ReplaceAll(stdin, delim, "")
	return "cat <<'" + delim + "' | " + line + "\n" + safe + "\n" + delim
}

// StateMatches reports whether candidate is one of allowed; an empty allowed
// list matches every state.
func StateMatches(candidate RemoteSandboxState, allowed []RemoteSandboxState) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, state := range allowed {
		if candidate == state {
			return true
		}
	}
	return false
}
