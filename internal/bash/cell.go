package bash

import (
	"fmt"

	"litdoc/internal"
)

type BashCell struct {
	block  internal.Block
	output internal.Output
}

func MakeBashCellFromRaw(content, indent string, output internal.Output) BashCell {
	return BashCell{
		block:  internal.MakeBlockFromRaw(internal.BlockKindFencedCode, content, indent, false),
		output: output,
	}
}

func ParseBashCell(
	block internal.Block,
	following []internal.Block,
) (
	internal.Cell,
	int,
	error,
) {
	// todo: test this
	output, consumed, err := internal.OutputFromBlocks(block, following)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing output: %w", err)
	}
	return BashCell{block: block, output: output}, consumed, nil
}

func (c BashCell) Execute() (internal.Cell, error) {
	// todo: test this
	return BashCell{
		block:  c.block,
		output: internal.MakeOutput("output", c.block.Indent()),
	}, nil
}

func (c BashCell) Render() (string, error) {
	// todo: test this
	return c.block.Render() + c.output.Render(), nil
}
