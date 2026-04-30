package internal_test

import (
	"strings"

	"litdoc/internal"
)

func text(indent, content string, continuation bool) internal.Block {
	return internal.MakeBlockFromRaw(internal.BlockKindText, content, indent, continuation)
}

func code(indent, content string, continuation bool) internal.Block {
	return internal.MakeBlockFromRaw(internal.BlockKindFencedCode, content, indent, continuation)
}

func cmnt(indent, content string, continuation bool) internal.Block {
	return internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, content, indent, continuation)
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
