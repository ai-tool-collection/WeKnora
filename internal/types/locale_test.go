package types

import "testing"

func TestNormalizeSupportedLocale(t *testing.T) {
	for _, locale := range []string{"en-US", "ko-KR", "ja-JP", "ru-RU"} {
		if got := NormalizeSupportedLocale(locale); got != locale {
			t.Errorf("NormalizeSupportedLocale(%q) = %q", locale, got)
		}
	}
	if got := NormalizeSupportedLocale("zh-CN"); got != "" {
		t.Errorf("NormalizeSupportedLocale(zh-CN) = %q, want empty", got)
	}
	if got := NormalizeSupportedLocale("fr-FR"); got != "" {
		t.Errorf("NormalizeSupportedLocale(unsupported) = %q, want empty", got)
	}
}
