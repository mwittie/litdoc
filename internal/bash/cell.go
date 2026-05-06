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

func MakeCellFromRaw(
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

func ParseCell(
	block internal.Block,
	following []internal.Block,
) (
	internal.Cell,
	int,
	error,
) {
	return parseCellWith(block, following, internal.OutputFromBlocks, Runner{})
}

func parseCellWith(
	block internal.Block,
	following []internal.Block,
	parseOutput func(internal.Block, []internal.Block) (internal.Output, int, error),
	runner internal.Runner,
) (
	internal.Cell,
	int,
	error,
) {
	output, consumed, err := parseOutput(block, following)
	if err != nil {
		return nil, 0, fmt.Errorf("parsing output: %w", err)
	}
	return Cell{block: block, output: output, runner: runner}, consumed, nil
}

func (c Cell) Execute() (internal.Cell, error) {
	stdout, stderr, exitCode, err := c.runner.Run(codeBody(c.block.Content()))
	if err != nil {
		return nil, err
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

// codeBody extracts the script lines from a fenced code block's content,
// stripping the opening and closing fence lines.
func codeBody(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return ""
	}
	return strings.Join(lines[1:len(lines)-2], "\n") + "\n"
}
