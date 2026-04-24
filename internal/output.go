package internal

import (
	"bytes"
	"strings"
)

const (
	outputBeginMarker = "<!-- BEGIN litdoc OUTPUT -->"
	outputEndMarker   = "<!-- END litdoc OUTPUT -->"
)

type Output struct {
	content string
}

func MakeOutput(content string) Output {
	return Output{content: content}
}

func (o Output) Render() string {
	if o.content == "" {
		return ""
	}
	return "\n" + outputBeginMarker + "\n" + o.content + "\n" + outputEndMarker + "\n"
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment && bytes.HasPrefix(b.content, []byte(outputBeginMarker))
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment && bytes.HasPrefix(b.content, []byte(outputEndMarker))
}

// OutputFromBlocks looks for an output block at the start of blocks,
// skipping leading whitespace text blocks. Returns the Output and the number
// of blocks consumed. consumed is 0 if no output block was found.
func OutputFromBlocks(blocks []Block) (Output, int) {
	i := 0
	for i < len(blocks) && blocks[i].kind == BlockKindText && strings.TrimSpace(string(blocks[i].content)) == "" {
		i++
	}
	if i >= len(blocks) || !isOutputBegin(blocks[i]) {
		return Output{}, 0
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

	return MakeOutput(strings.TrimSuffix(buf.String(), "\n")), i
}
