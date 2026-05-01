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
	assert.Contains(t, got.Render(""), content)
}

func TestOutput_RenderWithIndent(t *testing.T) {
	// given
	content := "hello"
	indent := "  "

	// when
	got := internal.MakeOutput(content)

	// then
	assert.Contains(t, got.Render(indent), indent+content)
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
			joinLines(
				"",
				internal.OutputBeginMarker,
				"hello",
				internal.OutputEndMarker+"\n",
			),
		},
		{
			"ensure content rendered with trailing newline",
			"hello",
			"",
			joinLines(
				"",
				internal.OutputBeginMarker,
				"hello",
				internal.OutputEndMarker+"\n",
			),
		},
		{
			"multiline content",
			"hello\nworld",
			"",
			joinLines(
				"",
				internal.OutputBeginMarker,
				"hello",
				"world",
				internal.OutputEndMarker+"\n",
			),
		},
		{
			"indent content",
			"hello\n",
			"  ",
			joinLines(
				"",
				"  "+internal.OutputBeginMarker,
				"  hello",
				"  "+internal.OutputEndMarker+"\n",
			),
		},
		{
			"blockquote content",
			"hello\n",
			"> ",
			joinLines(
				">",
				"> "+internal.OutputBeginMarker,
				"> hello",
				"> "+internal.OutputEndMarker+"\n",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			output := internal.MakeOutput(tt.content)

			// when
			got := output.Render(tt.indent)

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOutputFromBlocks(t *testing.T) {
	wantOutput := func(indent, content string) string {
		return internal.MakeOutput(content).Render(indent)
	}

	t.Run("output block is scanned in", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			cmnt("", internal.OutputBeginMarker, false),
			text("", "hello\n", false),
			cmnt("", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("", "hello"), output.Render(""))
		assert.Equal(t, 3, consumed)
	})

	t.Run("leading whitespace blocks are skipped", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			text("", "\n", false),
			cmnt("", internal.OutputBeginMarker, false),
			text("", "hello\n", false),
			cmnt("", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("", "hello"), output.Render(""))
		assert.Equal(t, 4, consumed)
	})

	t.Run("indented output block is scanned in", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			cmnt("    ", internal.OutputBeginMarker, false),
			text("    ", "hello\nworld\n", false),
			cmnt("    ", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("    ", "hello\nworld"), output.Render("    "))
		assert.Equal(t, 3, consumed)
	})

	t.Run("indented output content must match marker indent", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			cmnt("    ", internal.OutputBeginMarker, false),
			text("  ", "hello\n", false),
			cmnt("    ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(blocks)

		// then
		require.ErrorContains(t, err, "output content indentation")
	})

	t.Run("indented output end marker must match begin marker indent", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			cmnt("    ", internal.OutputBeginMarker, false),
			text("    ", "hello\n", false),
			cmnt("  ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(blocks)

		// then
		require.ErrorContains(t, err, "output end marker indentation")
	})

	t.Run("no output block returns zero value and zero consumed", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			text("", "some text\n", false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, "", output.Render(""))
		assert.Equal(t, 0, consumed)
	})

	t.Run("empty blocks returns zero value and zero consumed", func(t *testing.T) {
		// when
		output, consumed, err := internal.OutputFromBlocks(nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, "", output.Render(""))
		assert.Equal(t, 0, consumed)
	})

	t.Run("opening marker without closing marker returns error", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			cmnt("", internal.OutputBeginMarker, false),
			text("", "hello\n", false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(blocks)

		// then
		require.ErrorContains(t, err, "unclosed output block")
	})
}
