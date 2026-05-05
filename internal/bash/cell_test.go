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

func TestBashCell(t *testing.T) {
	t.Run("without output", func(t *testing.T) {
		// given
		code := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		cell := bash.MakeBashCellFromRaw(code, "", internal.Output{})

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
		output := internal.MakeOutput("hello")
		cell := bash.MakeBashCellFromRaw(fencedCode, "", output)

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, fencedCode+output.Render(""), gotContent)
	})

	t.Run("execute produces stub output", func(t *testing.T) {
		// given
		fencedCode := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		cell := bash.MakeBashCellFromRaw(fencedCode, "", internal.Output{})

		// when
		gotCell, err := cell.Execute()

		// then
		require.NoError(t, err)
		rendered, err := gotCell.Render()
		require.NoError(t, err)
		assert.Equal(t, fencedCode+internal.MakeOutput("output").Render(""), rendered)
	})
}
