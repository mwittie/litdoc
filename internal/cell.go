package internal

import (
	"fmt"
	"strings"
)

type Cell interface {
	Execute() (Cell, error)
	Render() (string, error)
}

type StaticCell struct {
	content string
}

func MakeStaticCellFromRaw(content string) StaticCell {
	return StaticCell{content: content}
}

func (t StaticCell) Execute() (Cell, error) {
	return t, nil
}

func (t StaticCell) Render() (string, error) {
	return t.content, nil
}

type BashCell struct {
	fencedCode string
	indent     string
	output     Output
}

func MakeBashCellFromRaw(fencedCode string, output Output) BashCell {
	return BashCell{
		fencedCode: fencedCode,
		indent:     leadingIndent(fencedCode),
		output:     output,
	}
}

func (c BashCell) Execute() (Cell, error) {
	return BashCell{
		fencedCode: c.fencedCode,
		indent:     c.indent,
		output:     MakeOutput("output"),
	}, nil
}

func (c BashCell) Render() (string, error) {
	return c.fencedCode + c.output.Render(renderIndent(c.indent)), nil
}

type InfoString struct {
	Lang   string
	Litdoc bool
}

func ParseInfoString(b Block) InfoString {
	firstLine := b.content
	if i := strings.IndexByte(b.content, '\n'); i >= 0 {
		firstLine = b.content[:i]
	}
	var raw string
	switch b.kind {
	case BlockKindFencedCode:
		raw = strings.TrimLeft(firstLine, "`~")
	case BlockKindHTMLComment:
		raw = strings.TrimSpace(strings.TrimPrefix(firstLine, "<!--"))
	default:
		return InfoString{}
	}
	parts := strings.SplitN(raw, " | ", 2)
	lang := strings.TrimSpace(parts[0])
	litdoc := len(parts) > 1 && strings.HasPrefix(strings.TrimSpace(parts[1]), "litdoc")
	return InfoString{Lang: lang, Litdoc: litdoc}
}

func Classify(blocks []Block) ([]Cell, error) {
	var cells []Cell
	i := 0
	for i < len(blocks) {
		b := blocks[i]
		switch b.kind {
		case BlockKindFencedCode, BlockKindHTMLComment:
			info := ParseInfoString(b)
			switch {
			case info.Litdoc && info.Lang == "bash":
				output, consumed, err := OutputFromBlocks(blocks[i+1:])
				if err != nil {
					return nil, fmt.Errorf("parsing output: %w", err)
				}
				cells = append(cells, BashCell{
					fencedCode: renderStaticBlock(b),
					indent:     b.indent,
					output:     output,
				})
				i += 1 + consumed
				continue
			case info.Litdoc:
				return nil, fmt.Errorf("unsupported language: %q", info.Lang)
			default:
				cells = append(cells, MakeStaticCellFromRaw(renderStaticBlock(b)))
			}
		default:
			cells = append(cells, MakeStaticCellFromRaw(renderStaticBlock(b)))
		}
		i++
	}
	return cells, nil
}

func renderStaticBlock(b Block) string {
	if len(b.indent) == 0 {
		return b.content
	}

	lines := strings.Split(b.content, "\n")
	var rendered strings.Builder
	renderedIndent := renderIndent(b.indent)
	blankLineIndent := renderBlankLineIndent(b.indent)
	for i, line := range lines {
		if i == len(lines)-1 && len(line) == 0 {
			break
		}
		if i > 0 {
			rendered.WriteByte('\n')
		}
		if len(line) == 0 {
			rendered.WriteString(blankLineIndent)
			continue
		}
		if i > 0 {
			rendered.WriteString(renderedIndent)
		} else if !b.continuation {
			rendered.WriteString(b.indent)
		}
		rendered.WriteString(line)
	}
	if strings.HasSuffix(b.content, "\n") {
		rendered.WriteByte('\n')
	}
	return rendered.String()
}

func renderBlankLineIndent(indent string) string {
	if idx := strings.LastIndex(indent, ">"); idx >= 0 {
		return indent[:idx+1]
	}
	return ""
}

func renderIndent(indent string) string {
	if idx := strings.LastIndex(indent, "> "); idx >= 0 {
		prefixLen := idx + len("> ")
		return indent[:prefixLen] + strings.Repeat(" ", len(indent)-prefixLen)
	}
	return strings.Repeat(" ", len(indent))
}

func leadingIndent(s string) string {
	line := s
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		line = s[:i]
	}
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return line[:i]
}

func Execute(cells []Cell) ([]Cell, error) {
	var executedCells []Cell
	for _, c := range cells {
		executed, err := c.Execute()
		if err != nil {
			return nil, fmt.Errorf("executing cell: %w", err)
		}
		executedCells = append(executedCells, executed)
	}
	return executedCells, nil
}

func Compose(cells []Cell) (string, error) {
	var dst strings.Builder
	for _, c := range cells {
		rendered, err := c.Render()
		if err != nil {
			return "", fmt.Errorf("rendering cell: %w", err)
		}
		dst.WriteString(rendered)
	}
	return dst.String(), nil
}
