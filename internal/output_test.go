package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeOutput(t *testing.T) {
	// given
	content := "hello"

	// when
	got := internal.MakeOutput(content)

	// then
	assert.Contains(t, got.Render(), content)
}

func TestOutput_WithIndent(t *testing.T) {
	// given
	content := "hello"
	indent := "  "

	// when
	got := internal.MakeOutput(content).WithIndent(indent)

	// then
	assert.Contains(t, got.Render(), indent+content)
}

func TestOutputRender(t *testing.T) {
	t.Run("empty output renders empty string", func(t *testing.T) {
		// given
		output := internal.Output{}

		// when
		got := output.Render()

		// then
		assert.Equal(t, "", got)
	})

	t.Run("non-empty output renders wrapped content", func(t *testing.T) {
		// given
		output := internal.MakeOutput("hello")

		// when
		got := output.Render()

		// then
		assert.Equal(t, "\n"+internal.OutputBeginMarker+"hello\n"+internal.OutputEndMarker, got)
	})
}

func TestOutputFromBlocks(t *testing.T) {
	htmlComment := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, []byte(content), "")
	}
	text := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindText, []byte(content), "")
	}
	wantOutput := func(content string) string {
		return internal.MakeOutput(content).Render()
	}

	t.Run("output block is scanned in", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			htmlComment(internal.OutputBeginMarker),
			text("hello\n"),
			htmlComment(internal.OutputEndMarker),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("hello"), output.Render())
		assert.Equal(t, 3, consumed)
	})

	t.Run("leading whitespace blocks are skipped", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			text("\n"),
			htmlComment(internal.OutputBeginMarker),
			text("hello\n"),
			htmlComment(internal.OutputEndMarker),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("hello"), output.Render())
		assert.Equal(t, 4, consumed)
	})

	t.Run("no output block returns zero value and zero consumed", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			text("some text\n"),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, "", output.Render())
		assert.Equal(t, 0, consumed)
	})

	t.Run("empty blocks returns zero value and zero consumed", func(t *testing.T) {
		// when
		output, consumed, err := internal.OutputFromBlocks(nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, "", output.Render())
		assert.Equal(t, 0, consumed)
	})

	t.Run("opening marker without closing marker returns error", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			htmlComment(internal.OutputBeginMarker),
			text("hello\n"),
		}

		// when
		_, _, err := internal.OutputFromBlocks(blocks)

		// then
		require.ErrorContains(t, err, "unclosed output block")
	})
}
