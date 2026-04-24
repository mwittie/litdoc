package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, "\n<!-- BEGIN litdoc OUTPUT -->\nhello\n<!-- END litdoc OUTPUT -->\n", got)
	})
}

func TestOutputFromBlocks(t *testing.T) {
	htmlComment := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, []byte(content))
	}
	text := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindText, []byte(content))
	}
	wantOutput := func(content string) string {
		return "\n<!-- BEGIN litdoc OUTPUT -->\n" + content + "\n<!-- END litdoc OUTPUT -->\n"
	}

	t.Run("output block is scanned in", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			htmlComment("<!-- BEGIN litdoc OUTPUT -->\n"),
			text("hello\n"),
			htmlComment("<!-- END litdoc OUTPUT -->\n"),
		}

		// when
		output, consumed := internal.OutputFromBlocks(blocks)

		// then
		assert.Equal(t, wantOutput("hello"), output.Render())
		assert.Equal(t, 3, consumed)
	})

	t.Run("leading whitespace blocks are skipped", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			text("\n"),
			htmlComment("<!-- BEGIN litdoc OUTPUT -->\n"),
			text("hello\n"),
			htmlComment("<!-- END litdoc OUTPUT -->\n"),
		}

		// when
		output, consumed := internal.OutputFromBlocks(blocks)

		// then
		assert.Equal(t, wantOutput("hello"), output.Render())
		assert.Equal(t, 4, consumed)
	})

	t.Run("no output block returns zero value and zero consumed", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			text("some text\n"),
		}

		// when
		output, consumed := internal.OutputFromBlocks(blocks)

		// then
		assert.Equal(t, "", output.Render())
		assert.Equal(t, 0, consumed)
	})

	t.Run("empty blocks returns zero value and zero consumed", func(t *testing.T) {
		// when
		output, consumed := internal.OutputFromBlocks(nil)

		// then
		assert.Equal(t, "", output.Render())
		assert.Equal(t, 0, consumed)
	})
}
