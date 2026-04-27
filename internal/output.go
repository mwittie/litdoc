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

	o.content = strings.TrimSuffix(o.content, "\n")

	if o.indent == "" {
		return "\n" + OutputBeginMarker + o.content + "\n" + OutputEndMarker
	}
	var buf strings.Builder
	buf.WriteString("\n" + o.indent + OutputBeginMarker)
	for _, line := range strings.Split(o.content, "\n") {
		buf.WriteString(o.indent + line + "\n")
	}
	buf.WriteString(o.indent + OutputEndMarker)
	return buf.String()
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		bytes.HasPrefix(b.content, []byte(OutputBeginMarker))
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		bytes.HasPrefix(b.content, []byte(OutputEndMarker))
}

// todo: make sure that it handles indent
func OutputFromBlocks(blocks []Block) (Output, int, error) {
	i := 0
	// skip empty text blocks
	for i < len(blocks) &&
		blocks[i].kind == BlockKindText &&
		strings.TrimSpace(string(blocks[i].content)) == "" {
		i++
	}

	// advance past a 'begin' block, or exit
	if i < len(blocks) && isOutputBegin(blocks[i]) {
		i++
	} else {
		return Output{}, 0, nil
	}

	// accumulate text until an 'end' block and return
	var buf strings.Builder
	for i < len(blocks) {
		if isOutputEnd(blocks[i]) {
			i++
			return MakeOutput(buf.String()), i, nil
		}
		buf.Write(blocks[i].content)
		i++
	}

	// report not finding an 'end' block
	return Output{}, 0, fmt.Errorf("unclosed output block: missing %q", OutputEndMarker)
}
