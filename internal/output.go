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

func (o Output) Render(linePrefix string) string {
	if o.content == "" {
		return ""
	}

	var buf strings.Builder
	buf.WriteString(blankBlockQuoteLinePrefix(linePrefix))
	buf.WriteByte('\n')
	buf.WriteString(linePrefix + OutputBeginMarker + "\n")
	for _, line := range strings.Split(strings.TrimSuffix(o.content, "\n"), "\n") {
		buf.WriteString(linePrefix + line + "\n")
	}
	buf.WriteString(linePrefix + OutputEndMarker + "\n")

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
	// Inline HTML comments can leave a whitespace-only continuation block for
	// the rest of the marker line. It is parser residue, not output content.
	i = skipOutputMarkerLineRemainder(blocks, i)

	var buf strings.Builder
	for i < len(blocks) {
		// A space before an inline-split end marker is emitted as a separate
		// text block, not as marker indentation. Drop it before matching END.
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
			// Consume the newline or trailing whitespace after an inline end
			// marker so callers resume at the next meaningful block.
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
