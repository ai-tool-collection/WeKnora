package handler

import (
	"strings"
	"testing"
)

// The channel-creation endpoint rejects any platform without a registered
// adapter factory, so this set must track the factories wired in the container.
// Removed China-hosted platforms must stay rejected.
func TestValidIMPlatforms_OnlyRetainedPlatforms(t *testing.T) {
	for _, platform := range []string{"slack", "telegram", "mattermost"} {
		if !validIMPlatforms[platform] {
			t.Errorf("platform %q is not accepted", platform)
		}
	}
	for _, platform := range []string{"wecom", "feishu", "lark", "dingtalk", "wechat", "qqbot", "yunzhijia", "nonsense"} {
		if validIMPlatforms[platform] {
			t.Errorf("platform %q is accepted", platform)
		}
	}
}

// The 400 message is derived from validIMPlatforms; it must not drift as
// platforms are added.
func TestInvalidIMPlatformError_ListsEveryPlatform(t *testing.T) {
	for platform := range validIMPlatforms {
		if !strings.Contains(invalidIMPlatformError, "'"+platform+"'") {
			t.Errorf("error message omits %q: %s", platform, invalidIMPlatformError)
		}
	}
}
