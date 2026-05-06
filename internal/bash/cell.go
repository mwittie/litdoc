package bash

import (
	"fmt"

	"litdoc/internal"
)

type Cell struct {
	block  internal.Block
	output internal.Output
}

func MakeCellFromRaw(content, indent string, output internal.Output) Cell {
	return Cell{
		block:  internal.MakeBlockFromRaw(internal.BlockKindFencedCode, content, indent, false),
		output: output,
	}
}

func ParseCell(
	block internal.Block,
	following []internal.Block,
) (
	internal.Cell,
	int,
	error,
) {
	return parseCellWith(block, following, internal.OutputFromBlocks)
}

func parseCellWith(
	block internal.Block,
	following []internal.Block,
	parseOutput func(internal.Block, []internal.Block) (internal.Output, int, error),
) (
	internal.Cell,
	int,
	error,
) {
	output, consumed, err := parseOutput(block, following)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing output: %w", err)
	}
	return Cell{block: block, output: output}, consumed, nil
}

func (c Cell) Execute() (internal.Cell, error) {
	return Cell{
		block:  c.block,
		output: internal.MakeOutput("output", c.block.Indent()),
	}, nil
}

func (c Cell) Render() (string, error) {
	return c.block.Render() + c.output.Render(), nil
}
