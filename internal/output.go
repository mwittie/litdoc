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
		return "\n" + OutputBeginMarker + o.content + "\n" + OutputEndMarker
	}
	var buf strings.Builder
	buf.WriteString("\n" + indent + OutputBeginMarker)
	for _, line := range strings.Split(o.content, "\n") {
		buf.WriteString(indent + line + "\n")
	}
	buf.WriteString(indent + OutputEndMarker)
	return buf.String()
}

func isOutputBegin(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		bytes.HasPrefix(bytes.TrimLeft(b.content, " \t"), []byte(OutputBeginMarker))
}

func isOutputEnd(b Block) bool {
	return b.kind == BlockKindHTMLComment &&
		bytes.HasPrefix(bytes.TrimLeft(b.content, " \t"), []byte(OutputEndMarker))
}

func blockLineIndent(b Block) string {
	line := b.content
	if i := bytes.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return string(line[:i])
}

func OutputFromBlocks(blocks []Block) (Output, string, int, error) {
	i := 0
	// skip empty text blocks
	for i < len(blocks) &&
		blocks[i].kind == BlockKindText &&
		strings.TrimSpace(string(blocks[i].content)) == "" {
		i++
	}

	// advance past a 'begin' block, or exit
	indent := ""
	if i < len(blocks) && isOutputBegin(blocks[i]) {
		indent = blockLineIndent(blocks[i])
		i++
	} else {
		return Output{}, "", 0, nil
	}

	// accumulate text until an 'end' block and return
	var buf strings.Builder
	for i < len(blocks) {
		if isOutputEnd(blocks[i]) {
			if got := blockLineIndent(blocks[i]); got != indent {
				return Output{}, "", 0, fmt.Errorf(
					"output end marker indentation: got %q, want %q",
					got,
					indent,
				)
			}
			i++
			return MakeOutput(buf.String()), indent, i, nil
		}
		unindented, err := unindentOutputContent(blocks[i].content, indent)
		if err != nil {
			return Output{}, "", 0, err
		}
		buf.Write(unindented)
		i++
	}

	// report not finding an 'end' block
	return Output{}, "", 0, fmt.Errorf("unclosed output block: missing %q", OutputEndMarker)
}

func unindentOutputContent(content []byte, indent string) ([]byte, error) {
	if indent == "" || len(content) == 0 {
		return content, nil
	}

	lines := bytes.SplitAfter(content, []byte("\n"))
	var buf bytes.Buffer
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		trimmedLine := bytes.TrimSuffix(line, []byte("\n"))
		if len(trimmedLine) == 0 {
			buf.Write(line)
			continue
		}
		if !bytes.HasPrefix(line, []byte(indent)) {
			return nil, fmt.Errorf("output content indentation: got %q, want prefix %q", string(line), indent)
		}
		buf.Write(line[len(indent):])
	}
	return buf.Bytes(), nil
}
