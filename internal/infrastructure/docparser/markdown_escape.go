package docparser

import "regexp"

// escapedImageSyntaxPattern matches markdown image references whose `[` was
// over-escaped to `\[` by html-to-markdown. The URL group mirrors the
// downstream image-extraction regex so escapes are only stripped for actual
// image references.
var escapedImageSyntaxPattern = regexp.MustCompile(`!\\\[(.*?)\\?\]\(([^()\n]*(?:\([^)]*\)[^()\n]*)*)\)`)

// unescapeMarkdownImageSyntax restores `![alt](url)` from html-to-markdown's
// over-escaped `!\[alt\](url)` form. Without this, the downstream image regex
// in ImageResolver fails to match and images are never persisted.
func unescapeMarkdownImageSyntax(content string) string {
	return escapedImageSyntaxPattern.ReplaceAllString(content, "![$1]($2)")
}
