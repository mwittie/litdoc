package bash_test

import (
	"strings"
	"testing"

	"litdoc/internal"
	"litdoc/internal/bash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestParser_Parse(t *testing.T) {
	block := internal.MakeBlockFromRaw(internal.BlockKindFencedCode, joinLines(
		"```bash",
		"echo hello",
		"```",
		"",
	), "", false)

	t.Run("assembles cell from block and output", func(t *testing.T) {
		// given
		output := internal.MakeOutput("hello", "")
		outputParser := NewMockOutputParser(t)
		outputParser.EXPECT().
			Parse(block, []internal.Block(nil)).
			Return(output, 3, nil)
		parser := bash.NewParserWith(nil, outputParser)

		// when
		cell, consumed, err := parser.Parse(block, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, 3, consumed)
		rendered, err := cell.Render()
		require.NoError(t, err)
		assert.Equal(t, block.Render()+output.Render(), rendered)
	})

	t.Run("output parsing error is wrapped", func(t *testing.T) {
		// given
		outputParser := NewMockOutputParser(t)
		outputParser.EXPECT().
			Parse(block, []internal.Block(nil)).
			Return(internal.Output{}, 0, assert.AnError)
		parser := bash.NewParserWith(nil, outputParser)

		// when
		_, _, err := parser.Parse(block, nil)

		// then
		require.ErrorContains(t, err, "parsing output")
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestCell_Execute(t *testing.T) {
	fencedCode := joinLines(
		"```bash",
		"echo hello",
		"```",
		"",
	)

	t.Run("success", func(t *testing.T) {
		// given
		runner := NewMockRunner(t)
		runner.EXPECT().
			Run("echo hello\n").
			Return("hello\n", "", 0, nil)
		cell := bash.MakeCellFromRaw(fencedCode, "", internal.MakeOutput("", ""), runner)

		// when
		gotCell, err := cell.Execute()

		// then
		require.NoError(t, err)
		rendered, err := gotCell.Render()
		require.NoError(t, err)
		assert.Equal(t, fencedCode+internal.MakeOutput("hello\n", "").Render(), rendered)
	})

	t.Run("non-zero exit code", func(t *testing.T) {
		// given
		runner := NewMockRunner(t)
		runner.EXPECT().
			Run("echo hello\n").
			Return("", "bash: command not found\n", 127, nil)
		cell := bash.MakeCellFromRaw(fencedCode, "", internal.MakeOutput("", ""), runner)

		// when
		_, err := cell.Execute()

		// then
		require.ErrorContains(t, err, "exit status 127")
		require.ErrorContains(t, err, "bash: command not found")
	})

	t.Run("exec error", func(t *testing.T) {
		// given
		runner := NewMockRunner(t)
		runner.EXPECT().
			Run("echo hello\n").
			Return("", "", 0, assert.AnError)
		cell := bash.MakeCellFromRaw(fencedCode, "", internal.MakeOutput("", ""), runner)

		// when
		_, err := cell.Execute()

		// then
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestCell_Render(t *testing.T) {
	t.Run("without output", func(t *testing.T) {
		// given
		code := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		cell := bash.MakeCellFromRaw(code, "", internal.MakeOutput("", ""), nil)

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, code, gotContent)
	})

	t.Run("with output", func(t *testing.T) {
		// given
		fencedCode := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		output := internal.MakeOutput("hello", "")
		cell := bash.MakeCellFromRaw(fencedCode, "", output, nil)

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, fencedCode+output.Render(), gotContent)
	})
}
