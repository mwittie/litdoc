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
	block Block
}

func MakeOutput(content, indent string) Output {
	if content == "" {
		return Output{}
	}
	linePrefix := RenderIndent(indent)
	assembled := "\n" + OutputBeginMarker + "\n" +
		strings.TrimSuffix(content, "\n") + "\n" +
		OutputEndMarker + "\n"
	return Output{block: MakeBlockFromRaw(BlockKindText, assembled, linePrefix, false)}
}

func (o Output) Render() string {
	return o.block.Render()
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		strings.HasPrefix(b.content, OutputBeginMarker)
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		strings.HasPrefix(b.content, OutputEndMarker)
}

func OutputFromBlocks(litdoc Block, blocks []Block) (Output, int, error) {
	// Output belongs to the preceding litdoc block, so its lines must use the
	// indentation that rendering that block would use for following content.
	outputIndent := RenderIndent(litdoc.indent)
	i := 0

	var err error
	i, err = skipWhitespaceLines(blocks, i, litdoc.indent, outputIndent)
	if err != nil {
		return Output{}, 0, err
	}

	beginLinePrefix := false
	if isWhitespaceBeforeOutputBegin(blocks, i) {
		if err := validateInlineMarkerPrefix(blocks[i], outputIndent, "begin"); err != nil {
			return Output{}, 0, err
		}
		beginLinePrefix = true
		i++
	}

	if i >= len(blocks) || !isOutputBegin(blocks[i]) {
		return Output{}, 0, nil
	}
	if err := validateOutputMarkerIndent(blocks[i], outputIndent, beginLinePrefix, "begin"); err != nil {
		return Output{}, 0, err
	}
	i++
	// Inline HTML comments can leave a whitespace-only continuation block for
	// the rest of the marker line. It is parser residue, not output content.
	i = skipOutputMarkerLineRemainder(blocks, i)

	var buf strings.Builder
	for i < len(blocks) {
		// A space before an inline-split end marker is emitted as a separate
		// text block, not as marker indentation. Validate and consume it before
		// matching END.
		endLinePrefix := false
		if isWhitespaceBeforeOutputEnd(blocks, i) {
			if err := validateInlineMarkerPrefix(blocks[i], outputIndent, "end"); err != nil {
				return Output{}, 0, err
			}
			endLinePrefix = true
			i++
		}
		if isOutputEnd(blocks[i]) {
			if err := validateOutputMarkerIndent(blocks[i], outputIndent, endLinePrefix, "end"); err != nil {
				return Output{}, 0, err
			}
			i++
			// Consume the newline or trailing whitespace after an inline end
			// marker so callers resume at the next meaningful block.
			i = skipOutputMarkerLineRemainder(blocks, i)
			return MakeOutput(buf.String(), litdoc.indent), i, nil
		}
		content, err := outputContent(blocks[i], outputIndent)
		if err != nil {
			return Output{}, 0, err
		}
		buf.WriteString(content)
		i++
	}

	return Output{}, 0, fmt.Errorf("unclosed output block: missing %q", OutputEndMarker)
}

func isWhitespaceText(b Block) bool {
	return b.kind == BlockKindText && strings.TrimSpace(b.content) == ""
}

// skipWhitespaceLines skips blank-line text blocks between a litdoc cell and
// its output block. Only blocks that contain a newline are skipped here;
// single-line whitespace-only blocks are inline marker prefixes emitted by
// the HTML comment splitter and are handled separately.
func skipWhitespaceLines(blocks []Block, i int, litdocIndent, outputIndent string) (int, error) {
	for i < len(blocks) &&
		isWhitespaceText(blocks[i]) &&
		strings.Contains(blocks[i].content, "\n") {
		if err := validateOutputBlankLineIndent(blocks[i], litdocIndent, outputIndent); err != nil {
			return 0, err
		}
		i++
	}
	return i, nil
}

func validateOutputBlankLineIndent(b Block, litdocIndent, outputIndent string) error {
	if litdocIndent == outputIndent || b.indent != litdocIndent {
		return nil
	}
	return fmt.Errorf(
		"output blank line indentation: got %q for content %q, want %q",
		b.indent,
		b.content,
		outputIndent,
	)
}

func isWhitespaceBeforeOutputBegin(blocks []Block, i int) bool {
	return i+1 < len(blocks) &&
		blocks[i].kind == BlockKindText &&
		!strings.Contains(blocks[i].content, "\n") &&
		strings.TrimSpace(blocks[i].content) == "" &&
		isOutputBegin(blocks[i+1])
}

func validateInlineMarkerPrefix(b Block, indent, marker string) error {
	if b.content == indent {
		return nil
	}
	return fmt.Errorf(
		"output %s marker indentation: got %q for content %q, want %q",
		marker,
		b.indent,
		b.content,
		indent,
	)
}

func validateOutputMarkerIndent(b Block, indent string, inlinePrefix bool, marker string) error {
	want := indent
	if inlinePrefix {
		want = ""
	}
	if b.indent == want {
		return nil
	}
	return fmt.Errorf(
		"output %s marker indentation: got %q for content %q, want %q",
		marker,
		b.indent,
		b.content,
		indent,
	)
}

func outputContent(b Block, indent string) (string, error) {
	if b.indent == indent {
		return b.content, nil
	}
	if b.indent == "" && isSpaceIndent([]byte(indent)) {
		if content, ok := stripLinePrefix(b.content, indent); ok {
			return content, nil
		}
	}
	return "", fmt.Errorf(
		"output content indentation: got %q for content %q, want %q",
		b.indent,
		b.content,
		indent,
	)
}

func stripLinePrefix(content, prefix string) (string, bool) {
	if prefix == "" {
		return content, true
	}

	var buf strings.Builder
	for len(content) > 0 {
		line := content
		if i := strings.IndexByte(content, '\n'); i >= 0 {
			line = content[:i+1]
		}
		if !strings.HasPrefix(line, prefix) {
			return "", false
		}
		buf.WriteString(strings.TrimPrefix(line, prefix))
		content = content[len(line):]
	}
	return buf.String(), true
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
