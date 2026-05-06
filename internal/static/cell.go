package static

import "litdoc/internal"

type Cell struct {
	block internal.Block
}

func makeCellFromRaw(content string) Cell {
	return Cell{block: internal.MakeBlockFromRaw(internal.BlockKindText, content, "", false)}
}

type Parser struct{}

func NewParser() Parser {
	return Parser{}
}

func (p Parser) Parse(
	block internal.Block,
	_ []internal.Block,
) (
	internal.Cell,
	int,
	error,
) {
	return Cell{block: block}, 0, nil
}

func (c Cell) Execute() (internal.Cell, error) {
	return c, nil
}

func (c Cell) Render() (string, error) {
	return c.block.Render(), nil
}
