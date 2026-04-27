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

func TestOutput_Render(t *testing.T) {
	tests := []struct {
		name    string
		content string
		indent  string
		want    string
	}{
		{
			"empty",
			"",
			"",
			"",
		},
		{
			"wrap content in markers",
			"hello\n",
			"",
			"\n" + internal.OutputBeginMarker + "hello\n" + internal.OutputEndMarker,
		},
		{
			"ensure content rendered with trailing newline",
			"hello",
			"",
			"\n" + internal.OutputBeginMarker + "hello\n" + internal.OutputEndMarker,
		},
		{
			"multiline content",
			"hello\nworld",
			"",
			"\n" + internal.OutputBeginMarker + "hello\nworld\n" + internal.OutputEndMarker,
		},
		{
			"indent content",
			"hello\n",
			"  ",
			"\n" +
				"  " + internal.OutputBeginMarker +
				"  " + "hello\n" +
				"  " + internal.OutputEndMarker,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			output := internal.MakeOutput(tt.content).WithIndent(tt.indent)

			// when
			got := output.Render()

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOutputFromBlocks(t *testing.T) {
	htmlComment := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, []byte(content))
	}
	text := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindText, []byte(content))
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

	t.Run("indented output block is scanned in", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			htmlComment("    " + internal.OutputBeginMarker),
			text("    hello\n    world\n"),
			htmlComment("    " + internal.OutputEndMarker),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, internal.MakeOutput("hello\nworld").WithIndent("    ").Render(), output.Render())
		assert.Equal(t, 3, consumed)
	})

	t.Run("indented output content must match marker indent", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			htmlComment("    " + internal.OutputBeginMarker),
			text("  hello\n"),
			htmlComment("    " + internal.OutputEndMarker),
		}

		// when
		_, _, err := internal.OutputFromBlocks(blocks)

		// then
		require.ErrorContains(t, err, "output content indentation")
	})

	t.Run("indented output end marker must match begin marker indent", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			htmlComment("    " + internal.OutputBeginMarker),
			text("    hello\n"),
			htmlComment("  " + internal.OutputEndMarker),
		}

		// when
		_, _, err := internal.OutputFromBlocks(blocks)

		// then
		require.ErrorContains(t, err, "output end marker indentation")
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
