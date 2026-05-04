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
	if blankLineIndent := renderBlankLineIndent(indent); blankLineIndent != "" {
		buf.WriteString(blankLineIndent)
		buf.WriteByte('\n')
	} else {
		buf.WriteByte('\n')
	}
	buf.WriteString(indent + OutputBeginMarker + "\n")
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
	i = skipOutputMarkerLineRemainder(blocks, i)

	var buf strings.Builder
	for i < len(blocks) {
		if isWhitespaceBeforeOutputEnd(blocks, i) {
			i++
			continue
		}
		if isOutputEnd(blocks[i]) {
			if blocks[i].indent != indent {
				return Output{}, 0, fmt.Errorf(
					"output end marker indentation: got %q for content %q, want %q",
					blocks[i].indent,
					blocks[i].content,
					indent,
				)
			}
			i++
			i = skipOutputMarkerLineRemainder(blocks, i)
			return MakeOutput(buf.String()), i, nil
		}
		if blocks[i].indent != indent {
			return Output{}, 0, fmt.Errorf(
				"output content indentation: got %q for content %q, want %q",
				blocks[i].indent,
				blocks[i].content,
				indent,
			)
		}
		buf.WriteString(blocks[i].content)
		i++
	}

	return Output{}, 0, fmt.Errorf("unclosed output block: missing %q", OutputEndMarker)
}

func skipOutputMarkerLineRemainder(blocks []Block, i int) int {
	if i < len(blocks) &&
		blocks[i].kind == BlockKindText &&
		blocks[i].continuation &&
		strings.TrimSpace(blocks[i].content) == "" {
		return i + 1
	}
	return i
}

func isWhitespaceBeforeOutputEnd(blocks []Block, i int) bool {
	return i+1 < len(blocks) &&
		blocks[i].kind == BlockKindText &&
		!strings.Contains(blocks[i].content, "\n") &&
		strings.TrimSpace(blocks[i].content) == "" &&
		isOutputEnd(blocks[i+1])
}
