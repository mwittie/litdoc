package static

import "litdoc/internal"

type StaticCell struct {
	content string
}

func MakeStaticCellFromRaw(content string) StaticCell {
	return StaticCell{content: content}
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
	return StaticCell{content: internal.RenderBlock(block)}, 0, nil
}

func (c StaticCell) Execute() (internal.Cell, error) {
	// todo: test this
	return c, nil
}

func (c StaticCell) Render() (string, error) {
	// todo: test this
	return c.content, nil
}
