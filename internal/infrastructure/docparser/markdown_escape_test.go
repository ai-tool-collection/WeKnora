package docparser

import "testing"

func TestUnescapeMarkdownImageSyntax(t *testing.T) {
	cases := map[string]string{
		`!\[chart\](images/a.png)`:         `![chart](images/a.png)`,
		`!\[x\](https://h/a_(1).png) tail`: `![x](https://h/a_(1).png) tail`,
		`text \[not an image\](link)`:      `text \[not an image\](link)`,
		`![already](clean.png)`:            `![already](clean.png)`,
	}
	for in, want := range cases {
		if got := unescapeMarkdownImageSyntax(in); got != want {
			t.Errorf("unescapeMarkdownImageSyntax(%q) = %q, want %q", in, got, want)
		}
	}
}
