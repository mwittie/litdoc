package internal

import (
	"bytes"
	"strings"
)

const (
	outputBeginMarker = "<!-- BEGIN litdoc OUTPUT -->"
	outputEndMarker   = "<!-- END litdoc OUTPUT -->"
)

func FormatOutput(output string) string {
	return outputBeginMarker + "\n" + output + "\n" + outputEndMarker + "\n"
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment && bytes.HasPrefix(b.content, []byte(outputBeginMarker))
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment && bytes.HasPrefix(b.content, []byte(outputEndMarker))
}

// ScanOutput looks for an output block at the start of blocks, skipping leading
// whitespace text blocks. Returns the raw output content and the number of
// blocks consumed. consumed is 0 if no output block was found.
func ScanOutput(blocks []Block) (output string, consumed int) {
	i := 0
	for i < len(blocks) && blocks[i].kind == BlockKindText && strings.TrimSpace(string(blocks[i].content)) == "" {
		i++
	}
	if i >= len(blocks) || !isOutputBegin(blocks[i]) {
		return "", 0
	}
	i++ // skip BEGIN block
	var buf strings.Builder
	for i < len(blocks) {
		if isOutputEnd(blocks[i]) {
			i++
			break
		}
		buf.Write(blocks[i].content)
		i++
	}
	return strings.TrimSuffix(buf.String(), "\n"), i
}
