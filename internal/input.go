package internal

import (
	"regexp"
	"strings"
)

// codeBodyRe strips the opening and closing delimiter lines from any block
// type. (?s) lets .* span newlines; greedy backtracking anchors the closing
// delimiter to the last non-empty line before end-of-string.
var codeBodyRe = regexp.MustCompile("(?s)^[^\n]*\n(.*)\n[^\n]+\n?$")

type InfoString struct {
	Lang   string
	Litdoc bool
}

func InfoStringFromBlock(b Block) InfoString {
	raw := b.headerLine()
	if raw == "" {
		return InfoString{}
	}
	parts := strings.SplitN(raw, " | ", 2)
	lang := strings.TrimSpace(parts[0])
	litdoc := len(parts) > 1 && strings.HasPrefix(strings.TrimSpace(parts[1]), "litdoc")
	return InfoString{Lang: lang, Litdoc: litdoc}
}

func CodeBody(content string) string {
	m := codeBodyRe.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	return m[1] + "\n"
}
