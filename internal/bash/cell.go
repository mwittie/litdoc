package bash

import (
	"fmt"

	"litdoc/internal"
)

type BashCell struct {
	content string
	indent  string
	output  internal.Output
}

func MakeBashCellFromRaw(content, indent string, output internal.Output) BashCell {
	return BashCell{
		content: content,
		indent:  indent,
		output:  output,
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
	return MakeBashCellFromRaw(block.Content(), block.Indent(), output), consumed, nil
}

func (c BashCell) Execute() (internal.Cell, error) {
	// todo: test this
	return BashCell{
		content: c.content,
		indent:  c.indent,
		output:  internal.MakeOutput("output"),
	}, nil
}

func (c BashCell) Render() (string, error) {
	// todo: test this
	fencedCode := internal.RenderContent(c.content, c.indent)
	return fencedCode + c.output.Render(internal.RenderIndent(c.indent)), nil
}
