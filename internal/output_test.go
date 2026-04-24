package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
)

func TestFormatOutput(t *testing.T) {
	got := internal.FormatOutput("hello")
	assert.Equal(t, "<!-- BEGIN litdoc OUTPUT -->\nhello\n<!-- END litdoc OUTPUT -->\n", got)
}

func TestScanOutput(t *testing.T) {
	htmlComment := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, []byte(content))
	}
	text := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindText, []byte(content))
	}

	t.Run("output block is scanned in", func(t *testing.T) {
		blocks := []internal.Block{
			htmlComment("<!-- BEGIN litdoc OUTPUT -->\n"),
			text("hello\n"),
			htmlComment("<!-- END litdoc OUTPUT -->\n"),
		}
		output, consumed := internal.ScanOutput(blocks)
		assert.Equal(t, "hello", output)
		assert.Equal(t, 3, consumed)
	})

	t.Run("leading whitespace blocks are skipped", func(t *testing.T) {
		blocks := []internal.Block{
			text("\n"),
			htmlComment("<!-- BEGIN litdoc OUTPUT -->\n"),
			text("hello\n"),
			htmlComment("<!-- END litdoc OUTPUT -->\n"),
		}
		output, consumed := internal.ScanOutput(blocks)
		assert.Equal(t, "hello", output)
		assert.Equal(t, 4, consumed)
	})

	t.Run("no output block returns empty and zero consumed", func(t *testing.T) {
		blocks := []internal.Block{
			text("some text\n"),
		}
		output, consumed := internal.ScanOutput(blocks)
		assert.Equal(t, "", output)
		assert.Equal(t, 0, consumed)
	})

	t.Run("empty blocks returns empty and zero consumed", func(t *testing.T) {
		output, consumed := internal.ScanOutput(nil)
		assert.Equal(t, "", output)
		assert.Equal(t, 0, consumed)
	})
}
