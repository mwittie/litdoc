package bash

import (
	"fmt"
	"strings"

	"litdoc/internal"
)

type Cell struct {
	block  internal.Block
	output internal.Output
	runner internal.Runner
}

func makeCellFromRaw(
	content, indent string,
	output internal.Output,
	runner internal.Runner,
) Cell {
	return Cell{
		block:  internal.MakeBlockFromRaw(internal.BlockKindFencedCode, content, indent, false),
		output: output,
		runner: runner,
	}
}

type Parser struct {
	runner       internal.Runner
	outputParser internal.OutputParser
}

func MakeParser() Parser {
	return makeParserFromRaw(Runner{}, internal.OutputParserFunc(internal.OutputFromBlocks))
}

func makeParserFromRaw(
	runner internal.Runner,
	outputParser internal.OutputParser,
) Parser {
	return Parser{runner: runner, outputParser: outputParser}
}

func (p Parser) Parse(
	block internal.Block,
	following []internal.Block,
) (
	internal.Cell,
	int,
	error,
) {
	output, consumed, err := p.outputParser.Parse(block, following)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing output: %w", err)
	}
	return Cell{block: block, output: output, runner: p.runner}, consumed, nil
}

func (c Cell) Execute() (internal.Cell, error) {
	stdout, stderr, exitCode, err := c.runner.Run(internal.CodeBody(c.block.Content()))
	if err != nil {
		return nil, fmt.Errorf("running cell: %w", err)
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("exit status %d: %s", exitCode, strings.TrimSpace(stderr))
	}
	return Cell{
		block:  c.block,
		runner: c.runner,
		output: internal.MakeOutput(stdout, c.block.Indent()),
	}, nil
}

func (c Cell) Render() (string, error) {
	return c.block.Render() + c.output.Render(), nil
}

