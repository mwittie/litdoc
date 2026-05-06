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

func TestMakeCellFromRaw(t *testing.T) {
	// given
	code := joinLines(
		"```bash",
		"echo hello",
		"```",
		"",
	)
	output := internal.MakeOutput("hello", "")
	cell := bash.MakeCellFromRaw(code, "", output)

	// when
	got, err := cell.Render()

	// then
	require.NoError(t, err)
	assert.Equal(t, "```bash\necho hello\n```\n"+output.Render(), got)
}

func TestParseCellWith(t *testing.T) {
	block := internal.MakeBlockFromRaw(internal.BlockKindFencedCode, joinLines(
		"```bash",
		"echo hello",
		"```",
		"",
	), "", false)

	t.Run("assembles cell from block and output", func(t *testing.T) {
		// given
		output := internal.MakeOutput("hello", "")
		parseOutput := func(internal.Block, []internal.Block) (internal.Output, int, error) {
			return output, 3, nil
		}

		// when
		cell, consumed, err := bash.ParseCellWith(block, nil, parseOutput)

		// then
		require.NoError(t, err)
		assert.Equal(t, 3, consumed)
		rendered, err := cell.Render()
		require.NoError(t, err)
		assert.Equal(t, block.Render()+output.Render(), rendered)
	})

	t.Run("output parsing error is wrapped", func(t *testing.T) {
		// given
		parseOutput := func(internal.Block, []internal.Block) (internal.Output, int, error) {
			return internal.Output{}, 0, assert.AnError
		}

		// when
		_, _, err := bash.ParseCellWith(block, nil, parseOutput)

		// then
		require.ErrorContains(t, err, "parsing output")
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestRender(t *testing.T) {
	t.Run("without output", func(t *testing.T) {
		// given
		code := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		cell := bash.MakeCellFromRaw(code, "", internal.MakeOutput("", ""))

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
		cell := bash.MakeCellFromRaw(fencedCode, "", output)

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, fencedCode+output.Render(), gotContent)
	})
}

func TestExecute(t *testing.T) {
	// given
	fencedCode := joinLines(
		"```bash",
		"echo hello",
		"```",
		"",
	)
	cell := bash.MakeCellFromRaw(fencedCode, "", internal.MakeOutput("", ""))

	// when
	gotCell, err := cell.Execute()

	// then
	require.NoError(t, err)
	rendered, err := gotCell.Render()
	require.NoError(t, err)
	assert.Equal(t, fencedCode+internal.MakeOutput("output", "").Render(), rendered)
}
