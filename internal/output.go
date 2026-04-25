package internal

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	OutputBeginMarker = "<!-- BEGIN LITDOC OUTPUT -->\n"
	OutputEndMarker   = "<!-- END LITDOC OUTPUT -->\n"
)

type Output struct {
	content string
	indent  string
}

func MakeOutput(content string) Output {
	return Output{content: content}
}

func (o Output) WithIndent(indent string) Output {
	o.indent = indent
	return o
}

func (o Output) Render() string {
	if o.content == "" {
		return ""
	}
	ind := o.indent
	if ind == "" {
		return "\n" + OutputBeginMarker + o.content + "\n" + OutputEndMarker
	}
	var buf strings.Builder
	buf.WriteString("\n" + ind + OutputBeginMarker)
	for _, line := range strings.Split(o.content, "\n") {
		buf.WriteString(ind + line + "\n")
	}
	buf.WriteString(ind + OutputEndMarker)
	return buf.String()
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment && bytes.HasPrefix(b.content, []byte(OutputBeginMarker))
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment && bytes.HasPrefix(b.content, []byte(OutputEndMarker))
}

// OutputFromBlocks looks for an output block at the start of blocks,
// skipping leading whitespace text blocks. Returns the Output and the number
// of blocks consumed. consumed is 0 if no output block was found.
// Returns an error if an opening marker is found without a closing marker.
func OutputFromBlocks(blocks []Block) (Output, int, error) {
	i := 0
	for i < len(blocks) && blocks[i].kind == BlockKindText && strings.TrimSpace(string(blocks[i].content)) == "" {
		i++
	}
	if i >= len(blocks) || !isOutputBegin(blocks[i]) {
		return Output{}, 0, nil
	}
	i++ // skip BEGIN block
	var buf strings.Builder
	for i < len(blocks) {
		if isOutputEnd(blocks[i]) {
			i++
			return MakeOutput(strings.TrimSuffix(buf.String(), "\n")), i, nil
		}
		buf.Write(blocks[i].content)
		i++
	}
	return Output{}, 0, fmt.Errorf("unclosed output block: missing %q", OutputEndMarker)
}
