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

func MakeStaticCellFromRaw(raw string) StaticCell {
	return StaticCell{content: raw}
}

func (t StaticCell) Execute() (Cell, error) {
	return t, nil
}

func (t StaticCell) Render() (string, error) {
	return t.content, nil
}

type BashCell struct {
	fencedCode string
	output     string
}

func MakeBashCellFromRaw(fencedCode, output string) BashCell {
	return BashCell{fencedCode: fencedCode, output: output}
}

func (c BashCell) Execute() (Cell, error) {
	return c, nil
}

func (c BashCell) Render() (string, error) {
	if c.output == "" {
		return c.fencedCode, nil
	}
	return c.fencedCode + "\n" + c.output, nil
}

type InfoString struct {
	Lang     string
	IsLitdoc bool
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
	isLitdoc := len(parts) > 1 && strings.HasPrefix(strings.TrimSpace(parts[1]), "litdoc")
	return InfoString{Lang: lang, IsLitdoc: isLitdoc}
}

func Classify(blocks []Block) ([]Cell, error) {
	var cells []Cell
	for _, b := range blocks {
		info := ParseInfoString(b)
		switch {
		case info.IsLitdoc && info.Lang == "bash":
			cell := MakeBashCellFromRaw(b.content, "")
			cells = append(cells, cell)
		case info.IsLitdoc:
			return nil, fmt.Errorf("unsupported language: %q", info.Lang)
		default:
			cells = append(cells, MakeStaticCellFromRaw(renderStaticBlock(b)))
		}
	}
	return cells, nil
}

func renderStaticBlock(b Block) string {
	if len(b.indent) == 0 {
		return b.content
	}

	lines := strings.Split(b.content, "\n")
	var rendered strings.Builder
	continuationIndent := blockContinuationIndent(b.indent)
	for i, line := range lines {
		if i == len(lines)-1 && len(line) == 0 {
			break
		}
		if i > 0 {
			rendered.WriteByte('\n')
			if len(line) > 0 {
				rendered.WriteString(continuationIndent)
			}
		} else {
			if len(line) > 0 && !b.continuation {
				rendered.WriteString(b.indent)
			}
		}
		rendered.WriteString(line)
	}
	if strings.HasSuffix(b.content, "\n") {
		rendered.WriteByte('\n')
	}
	return rendered.String()
}

func blockContinuationIndent(indent string) string {
	if idx := strings.LastIndex(indent, "> "); idx >= 0 {
		prefixLen := idx + len("> ")
		return indent[:prefixLen] + strings.Repeat(" ", len(indent)-prefixLen)
	}
	return strings.Repeat(" ", len(indent))
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
