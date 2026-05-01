package internal

import (
	"fmt"
	"strings"
)

const (
	OutputBeginMarker = "<!-- BEGIN LITDOC OUTPUT -->"
	OutputEndMarker   = "<!-- END LITDOC OUTPUT -->"
)

type Output struct {
	content string
}

func MakeOutput(content string) Output {
	return Output{content: content}
}

func (o Output) Render(indent string) string {
	if o.content == "" {
		return ""
	}

	o.content = strings.TrimSuffix(o.content, "\n")

	if indent == "" {
		return "\n" + OutputBeginMarker + "\n" + o.content + "\n" + OutputEndMarker + "\n"
	}
	var buf strings.Builder
	buf.WriteString("\n" + indent + OutputBeginMarker + "\n")
	for _, line := range strings.Split(o.content, "\n") {
		buf.WriteString(indent + line + "\n")
	}
	buf.WriteString(indent + OutputEndMarker + "\n")
	return buf.String()
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		strings.HasPrefix(b.content, OutputBeginMarker)
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		strings.HasPrefix(b.content, OutputEndMarker)
}

func OutputFromBlocks(blocks []Block) (Output, int, error) {
	i := 0
	for i < len(blocks) &&
		blocks[i].kind == BlockKindText &&
		strings.TrimSpace(blocks[i].content) == "" {
		i++
	}

	if i >= len(blocks) || !isOutputBegin(blocks[i]) {
		return Output{}, 0, nil
	}
	indent := blocks[i].indent
	i++

	var buf strings.Builder
	for i < len(blocks) {
		if isOutputEnd(blocks[i]) {
			if blocks[i].indent != indent {
				return Output{}, 0, fmt.Errorf(
					"output end marker indentation: got %q, want %q",
					blocks[i].indent,
					indent,
				)
			}
			i++
			return MakeOutput(buf.String()), i, nil
		}
		if blocks[i].indent != indent {
			return Output{}, 0, fmt.Errorf(
				"output content indentation: got %q, want %q",
				blocks[i].indent,
				indent,
			)
		}
		buf.WriteString(blocks[i].content)
		i++
	}

	return Output{}, 0, fmt.Errorf("unclosed output block: missing %q", OutputEndMarker)
}
