package static

import "litdoc/internal"

type Cell struct {
	block internal.Block
}

func MakeCellFromRaw(content string) Cell {
	return Cell{block: internal.MakeBlockFromRaw(internal.BlockKindText, content, "", false)}
}

func ParseCell(
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
