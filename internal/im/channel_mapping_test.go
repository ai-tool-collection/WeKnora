package im

import (
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/types"
)

// Files uploaded through an IM channel are tagged with a knowledge Channel.
// Platforms without a dedicated channel, including removed China-hosted
// platforms still present in legacy rows, fall back to the generic one.
func TestIMPlatformToChannel(t *testing.T) {
	cases := map[string]string{
		"slack":      types.ChannelSlack,
		"Slack":      types.ChannelSlack, // matching is case-insensitive
		"telegram":   types.ChannelIM,
		"mattermost": types.ChannelIM,
		"feishu":     types.ChannelIM,
		"wecom":      types.ChannelIM,
		"":           types.ChannelIM,
	}

	for platform, want := range cases {
		if got := imPlatformToChannel(platform); got != want {
			t.Errorf("imPlatformToChannel(%q) = %q, want %q", platform, got, want)
		}
	}
}
