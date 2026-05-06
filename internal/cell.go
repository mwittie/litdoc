package internal

import (
	"fmt"
	"strings"
)

type Cell interface {
	Execute() (Cell, error)
	Render() (string, error)
}

type CellParser func(block Block, following []Block) (Cell, int, error)

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

func Classify(blocks []Block, parsers map[string]CellParser) ([]Cell, error) {
	// todo: make sure all this gets tested
	static, ok := parsers["static"]
	if !ok {
		return nil, fmt.Errorf("no static parser provided")
	}
	var cells []Cell
	i := 0
	for i < len(blocks) {
		b := blocks[i]
		switch b.kind {
		case BlockKindFencedCode, BlockKindHTMLComment:
			info := ParseInfoString(b)
			switch {
			case info.Litdoc:
				parser, ok := parsers[info.Lang]
				if !ok {
					return nil, fmt.Errorf("unsupported language: %q", info.Lang)
				}
				cell, consumed, err := parser(b, blocks[i+1:])
				if err != nil {
					return nil, err
				}
				cells = append(cells, cell)
				i += 1 + consumed
				continue
			default:
				cell, _, err := static(b, nil)
				if err != nil {
					return nil, err
				}
				cells = append(cells, cell)
			}
		default:
			cell, _, err := static(b, nil)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cell)
		}
		i++
	}
	return cells, nil
}

func blankBlockQuoteLinePrefix(indent string) string {
	if idx := strings.LastIndex(indent, ">"); idx >= 0 {
		return indent[:idx+1]
	}
	return ""
}

func RenderIndent(indent string) string {
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
