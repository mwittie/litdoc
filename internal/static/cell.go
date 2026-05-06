package static

import "litdoc/internal"

type StaticCell struct {
	block internal.Block
}

func MakeStaticCellFromRaw(content string) StaticCell {
	return StaticCell{block: internal.MakeBlockFromRaw(internal.BlockKindText, content, "", false)}
}

func ParseStaticCell(
	block internal.Block,
	_ []internal.Block,
) (
	internal.Cell,
	int,
	error,
) {
	// todo: test this
	return StaticCell{block: block}, 0, nil
}

func (c StaticCell) Execute() (internal.Cell, error) {
	// todo: test this
	return c, nil
}

func (c StaticCell) Render() (string, error) {
	// todo: test this
	return c.block.Render(), nil
}
