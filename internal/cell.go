package internal

import (
	"bytes"
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
	output     Output
}

func MakeBashCellFromRaw(fencedCode string, output Output) BashCell {
	return BashCell{fencedCode: fencedCode, output: output}
}

func (c BashCell) Execute() (Cell, error) {
	return BashCell{fencedCode: c.fencedCode, output: MakeOutput("output")}, nil // stub
}

func (c BashCell) Render() (string, error) {
	return c.fencedCode + c.output.Render(), nil
}

type InfoString struct {
	Lang     string
	IsLitdoc bool
}

func ParseInfoString(b Block) InfoString {
	firstLine := b.content
	if i := bytes.IndexByte(b.content, '\n'); i >= 0 {
		firstLine = b.content[:i]
	}
	var raw []byte
	switch b.kind {
	case BlockKindFencedCode:
		raw = bytes.TrimLeft(firstLine, "`~")
	case BlockKindHTMLComment:
		raw = bytes.TrimSpace(bytes.TrimPrefix(firstLine, []byte("<!--")))
	default:
		return InfoString{}
	}
	parts := bytes.SplitN(raw, []byte(" | "), 2)
	lang := string(bytes.TrimSpace(parts[0]))
	isLitdoc := len(parts) > 1 && bytes.HasPrefix(bytes.TrimSpace(parts[1]), []byte("litdoc"))
	return InfoString{Lang: lang, IsLitdoc: isLitdoc}
}

func BashCellFromBlocks(blocks []Block) (BashCell, int) {
	fencedCode := string(blocks[0].content)
	output, consumed := OutputFromBlocks(blocks[1:])
	return MakeBashCellFromRaw(fencedCode, output), 1 + consumed
}

func Classify(blocks []Block) ([]Cell, error) {
	var cells []Cell
	i := 0
	for i < len(blocks) {
		b := blocks[i]
		info := ParseInfoString(b)
		switch {
		case info.IsLitdoc && info.Lang == "bash":
			cell, consumed := BashCellFromBlocks(blocks[i:])
			cells = append(cells, cell)
			i += consumed
			continue
		case info.IsLitdoc:
			return nil, fmt.Errorf("unsupported language: %q", info.Lang)
		default:
			cells = append(cells, MakeStaticCellFromRaw(string(b.content)))
		}
		i++
	}
	return cells, nil
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
