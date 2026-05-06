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
	got := internal.MakeOutput(content, "")

	// then
	assert.Contains(t, got.Render(), content)
}

func TestOutput_RenderWithIndent(t *testing.T) {
	// given
	content := "hello"
	indent := "  "

	// when
	got := internal.MakeOutput(content, indent)

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
			name:    "empty",
			content: "",
			indent:  "",
			want:    "",
		},
		{
			name:    "wrap content in markers",
			content: "hello\n",
			indent:  "",
			want: joinLines(
				"",
				internal.OutputBeginMarker,
				"hello",
				internal.OutputEndMarker+"\n",
			),
		},
		{
			name:    "ensure content rendered with trailing newline",
			content: "hello",
			indent:  "",
			want: joinLines(
				"",
				internal.OutputBeginMarker,
				"hello",
				internal.OutputEndMarker+"\n",
			),
		},
		{
			name:    "multiline content",
			content: "hello\nworld",
			indent:  "",
			want: joinLines(
				"",
				internal.OutputBeginMarker,
				"hello",
				"world",
				internal.OutputEndMarker+"\n",
			),
		},
		{
			name:    "indent content",
			content: "hello\n",
			indent:  "  ",
			want: joinLines(
				"",
				"  "+internal.OutputBeginMarker,
				"  hello",
				"  "+internal.OutputEndMarker+"\n",
			),
		},
		{
			name:    "blockquote content",
			content: "hello\n",
			indent:  "> ",
			want: joinLines(
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
			output := internal.MakeOutput(tt.content, tt.indent)

			// when
			got := output.Render()

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOutputFromBlocks(t *testing.T) {
	wantOutput := func(indent, content string) string {
		return internal.MakeOutput(content, indent).Render()
	}

	t.Run("output block is scanned in", func(t *testing.T) {
		// given
		litdoc := code("", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("", internal.OutputBeginMarker, false),
			text("", "hello\n", false),
			cmnt("", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("", "hello"), output.Render())
		assert.Equal(t, 3, consumed)
	})

	t.Run("leading whitespace blocks are skipped", func(t *testing.T) {
		// given
		litdoc := code("", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			text("", "\n", false),
			cmnt("", internal.OutputBeginMarker, false),
			text("", "hello\n", false),
			cmnt("", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("", "hello"), output.Render())
		assert.Equal(t, 4, consumed)
	})

	t.Run("indented output block is scanned in", func(t *testing.T) {
		// given
		litdoc := code("    ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("    ", internal.OutputBeginMarker, false),
			text("    ", "hello\nworld\n", false),
			cmnt("    ", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("    ", "hello\nworld"), output.Render())
		assert.Equal(t, 3, consumed)
	})

	t.Run("inline marker line remainders are consumed", func(t *testing.T) {
		// given
		litdoc := code(" ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			text("", "\n", false),
			text("", " ", false),
			cmnt("", internal.OutputBeginMarker, false),
			text("", "\n", true),
			text("", " output\n", false),
			text("", " ", false),
			cmnt("", internal.OutputEndMarker, false),
			text("", "\n", true),
			text("", "\n", false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput(" ", "output"), output.Render())
		assert.Equal(t, 8, consumed)
	})

	t.Run("parser-space-prefixed output is normalized", func(t *testing.T) {
		// given
		litdoc := code("  ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			text("", "\n", false),
			text("", "  ", false),
			cmnt("", internal.OutputBeginMarker, false),
			text("", "\n", true),
			text("", "  hello\n  world\n", false),
			text("", "  ", false),
			cmnt("", internal.OutputEndMarker, false),
			text("", "\n", true),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("  ", "hello\nworld"), output.Render())
		assert.Equal(t, 8, consumed)
	})

	t.Run("list item output uses rendered indentation", func(t *testing.T) {
		// given
		litdoc := code("- ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			text("  ", "\n", false),
			cmnt("  ", internal.OutputBeginMarker, false),
			text("  ", "hello\n", false),
			cmnt("  ", internal.OutputEndMarker, false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, wantOutput("  ", "hello"), output.Render())
		assert.Equal(t, 4, consumed)
	})

	t.Run("list item output blank line must use rendered indentation", func(t *testing.T) {
		// given
		litdoc := code("- ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			text("- ", "\n", false),
			cmnt("  ", internal.OutputBeginMarker, false),
			text("  ", "hello\n", false),
			cmnt("  ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.ErrorContains(t, err, "output blank line indentation")
	})

	t.Run("indented output content must match marker indent", func(t *testing.T) {
		// given
		litdoc := code("    ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("    ", internal.OutputBeginMarker, false),
			text("  ", "hello\n", false),
			cmnt("    ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.ErrorContains(t, err, "output content indentation")
	})

	t.Run("indented output end marker must match begin marker indent", func(t *testing.T) {
		// given
		litdoc := code("    ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("    ", internal.OutputBeginMarker, false),
			text("    ", "hello\n", false),
			cmnt("  ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.ErrorContains(t, err, "output end marker indentation")
	})

	t.Run("output begin marker must match litdoc indent", func(t *testing.T) {
		// given
		litdoc := code("  ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("", internal.OutputBeginMarker, false),
			text("  ", "hello\n", false),
			cmnt("  ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.ErrorContains(t, err, "output begin marker indentation")
	})

	t.Run("list item output marker must use rendered indentation", func(t *testing.T) {
		// given
		litdoc := code("- ", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("- ", internal.OutputBeginMarker, false),
			text("  ", "hello\n", false),
			cmnt("  ", internal.OutputEndMarker, false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.ErrorContains(t, err, "output begin marker indentation")
	})

	t.Run("no output block returns zero value and zero consumed", func(t *testing.T) {
		// given
		litdoc := code("", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			text("", "some text\n", false),
		}

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.NoError(t, err)
		assert.Equal(t, "", output.Render())
		assert.Equal(t, 0, consumed)
	})

	t.Run("empty blocks returns zero value and zero consumed", func(t *testing.T) {
		// given
		litdoc := code("", "```bash | litdoc\n```\n", false)

		// when
		output, consumed, err := internal.OutputFromBlocks(litdoc, nil)

		// then
		require.NoError(t, err)
		assert.Equal(t, "", output.Render())
		assert.Equal(t, 0, consumed)
	})

	t.Run("opening marker without closing marker returns error", func(t *testing.T) {
		// given
		litdoc := code("", "```bash | litdoc\n```\n", false)
		blocks := []internal.Block{
			cmnt("", internal.OutputBeginMarker, false),
			text("", "hello\n", false),
		}

		// when
		_, _, err := internal.OutputFromBlocks(litdoc, blocks)

		// then
		require.ErrorContains(t, err, "unclosed output block")
	})
}
