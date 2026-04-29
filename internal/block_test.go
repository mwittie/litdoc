package internal_test

import (
	"strings"
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantBlock struct {
	kind    internal.BlockKind
	indent  string
	content string
}

func text(indent, content string) wantBlock {
	return wantBlock{internal.BlockKindText, indent, content}
}

func code(indent, content string) wantBlock {
	return wantBlock{internal.BlockKindFencedCode, indent, content}
}

func comment(indent, content string) wantBlock {
	return wantBlock{internal.BlockKindHTMLComment, indent, content}
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestMakeBlock(t *testing.T) {
	// given
	kind := internal.BlockKindFencedCode
	content := []byte("```bash\necho hello\n```\n")
	indent := []byte("> ")

	// when
	got := internal.MakeBlock(kind, content, indent)

	// then
	assert.Equal(t, kind, got.Kind())
	assert.Equal(t, content, got.Content())
	assert.Equal(t, indent, got.Indent())
}

func TestMakeBlocksFromMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []wantBlock
	}{
		// Text — baseline cases that should produce no FencedCode or HTMLComment blocks.
		{
			name:  "Text/heading",
			input: "# Hello\n",
			want: []wantBlock{
				text("", "# Hello\n"),
			},
		},
		{
			name: "Text/heading/nested",
			input: joinLines(
				"# Level 1",
				"",
				"some text",
				"",
				"## Level 2",
				"",
				"more text",
				"",
			),
			want: []wantBlock{
				text("", "# Level 1\n"),
				text("", "\n"),
				text("", "some text\n"),
				text("", "\n"),
				text("", "## Level 2\n"),
				text("", "\n"),
				text("", "more text\n"),
			},
		},
		{
			name: "Text/paragraph/grouped",
			input: joinLines(
				"text",
				"and more text",
				"",
				"last line",
			),
			want: []wantBlock{
				text("", "text\nand more text\n"),
				text("", "\n"),
				text("", "last line"),
			},
		},
		{
			name: "Text/paragraph/line-break",
			input: joinLines(
				"text",
				"and more text  ",
				"after line break",
			),
			want: []wantBlock{
				text("", "text\nand more text  \nafter line break"),
			},
		},
		{
			name: "Text/indented-code-as-text",
			input: joinLines(
				"    echo \"hello, \"",
				"    echo \"world\"",
			),
			want: []wantBlock{
				text("", "    echo \"hello, \"\n    echo \"world\""),
			},
		},
		{
			name:  "Text/inline-code-span",
			input: "text `echo \"hello, \" text`",
			want: []wantBlock{
				text("", "text `echo \"hello, \" text`"),
			},
		},
		{
			name:  "Text/html-block-not-comment",
			input: "<div>\nhello\n</div>\n",
			want: []wantBlock{
				text("", "<div>\nhello\n</div>\n"),
			},
		},
		{
			name: "Text/blockquote/multiline",
			input: joinLines(
				"> hello, ",
				"> world",
				"> ",
				"> again",
			),
			want: []wantBlock{
				text("> ", "hello, \nworld\n"),
				text("> ", "\n"),
				text("> ", "again"),
			},
		},

		// FencedCode — top-level (CommonMark §4.5).
		{
			name: "FencedCode/top-level/backtick",
			input: joinLines(
				"```bash",
				"echo \"hello\"",
				"```",
				"",
			),
			want: []wantBlock{
				code("", "```bash\necho \"hello\"\n```\n"),
			},
		},
		{
			name: "FencedCode/top-level/backtick-no-trailing-newline",
			input: joinLines(
				"```bash",
				"echo \"hello\"",
				"```",
			),
			want: []wantBlock{
				code("", "```bash\necho \"hello\"\n```"),
			},
		},
		{
			name: "FencedCode/top-level/tilde",
			input: joinLines(
				"~~~bash",
				"echo \"hello\"",
				"~~~",
			),
			want: []wantBlock{
				code("", "~~~bash\necho \"hello\"\n~~~"),
			},
		},
		{
			name: "FencedCode/top-level/longer-fence-around-fence",
			input: joinLines(
				"````md",
				"```bash",
				"echo \"hello\"",
				"```",
				"````",
			),
			want: []wantBlock{
				code("", "````md\n```bash\necho \"hello\"\n```\n````"),
			},
		},
		{
			name: "FencedCode/top-level/no-info-string",
			input: joinLines(
				"```",
				"echo \"hello\"",
				"```",
			),
			want: []wantBlock{
				code("", "```\necho \"hello\"\n```"),
			},
		},
		{
			name: "FencedCode/top-level/empty",
			input: joinLines(
				"```bash",
				"```",
			),
			want: []wantBlock{
				code("", "```bash\n```"),
			},
		},
		{
			name: "FencedCode/top-level/consecutive",
			input: joinLines(
				"```bash",
				"echo a",
				"```",
				"",
				"```bash",
				"echo b",
				"```",
			),
			want: []wantBlock{
				code("", "```bash\necho a\n```\n"),
				text("", "\n"),
				code("", "```bash\necho b\n```"),
			},
		},
		{
			name: "FencedCode/top-level/indented-1-space",
			input: joinLines(
				" ```bash",
				" echo \"hello\"",
				" ```",
			),
			want: []wantBlock{
				code(" ", "```bash\necho \"hello\"\n```"),
			},
		},
		{
			name: "FencedCode/top-level/indented-3-spaces",
			input: joinLines(
				"   ```bash",
				"   echo \"hello\"",
				"   ```",
			),
			want: []wantBlock{
				code("   ", "```bash\necho \"hello\"\n```"),
			},
		},
		{
			name: "FencedCode/top-level/unclosed-runs-to-eof",
			input: joinLines(
				"```bash",
				"echo \"hello\"",
			),
			want: []wantBlock{
				code("", "```bash\necho \"hello\""),
			},
		},

		// FencedCode — inside containers.
		{
			name: "FencedCode/blockquote",
			input: joinLines(
				"> text",
				"> ```bash",
				"> echo \"hello\"",
				"> ```",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				code("> ", "```bash\necho \"hello\"\n```\n"),
				text("> ", "text"),
			},
		},
		{
			name: "FencedCode/blockquote/nested",
			input: joinLines(
				"> text",
				"> > ```bash",
				"> > echo \"hello\"",
				"> > ```",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				code("> > ", "```bash\necho \"hello\"\n```\n"),
				text("> ", "text"),
			},
		},
		{
			name: "FencedCode/list/dash",
			input: joinLines(
				"- text",
				"- ```bash",
				"  echo \"hello\"",
				"  ```",
				"- text",
			),
			want: []wantBlock{
				text("- ", "text\n"),
				code("- ", "```bash\necho \"hello\"\n```\n"),
				text("- ", "text"),
			},
		},
		{
			name: "FencedCode/list/dash/nested",
			input: joinLines(
				"- text",
				"  - ```bash",
				"    echo \"hello\"",
				"    ```",
				"- text",
			),
			want: []wantBlock{
				text("- ", "text\n"),
				code("  - ", "```bash\necho \"hello\"\n```\n"),
				text("- ", "text"),
			},
		},
		{
			name: "FencedCode/list/plus-tilde-fence",
			input: joinLines(
				"+ text",
				"+ ~~~bash",
				"  echo \"hello\"",
				"  ~~~",
				"+ text",
			),
			want: []wantBlock{
				text("+ ", "text\n"),
				code("+ ", "~~~bash\necho \"hello\"\n~~~\n"),
				text("+ ", "text"),
			},
		},
		{
			name: "FencedCode/list/ordered",
			input: joinLines(
				"1. text",
				"2. ```bash",
				"   echo \"hello\"",
				"   ```",
				"3. text",
			),
			want: []wantBlock{
				text("1. ", "text\n"),
				code("2. ", "```bash\necho \"hello\"\n```\n"),
				text("3. ", "text"),
			},
		},
		{
			name: "FencedCode/list/ordered/nested",
			input: joinLines(
				"1. text",
				"   1. ```bash",
				"      echo \"hello\"",
				"      ```",
				"2. text",
			),
			want: []wantBlock{
				text("1. ", "text\n"),
				code("   1. ", "```bash\necho \"hello\"\n```\n"),
				text("2. ", "text"),
			},
		},
		{
			name: "FencedCode/list/blockquote",
			input: joinLines(
				"- text",
				"  > ```bash",
				"  > echo \"hello\"",
				"  > ```",
				"- text",
			),
			want: []wantBlock{
				text("- ", "text\n"),
				code("  > ", "```bash\necho \"hello\"\n```\n"),
				text("- ", "text"),
			},
		},
		{
			name: "FencedCode/blockquote/list-nested",
			input: joinLines(
				"> text",
				"> - item",
				">   - ```bash",
				">     echo \"hello\"",
				">     ```",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				text("> - ", "item\n"),
				code(">   - ", "```bash\necho \"hello\"\n```\n"),
				text("> ", "text"),
			},
		},
		{
			name: "FencedCode/blockquote/list-ordered",
			input: joinLines(
				"> text",
				"> 1. ```bash",
				">    echo \"hello\"",
				">    ```",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				code("> 1. ", "```bash\necho \"hello\"\n```\n"),
				text("> ", "text"),
			},
		},

		// HTMLComment — top-level (CommonMark §4.6 type 2).
		{
			name:  "HTMLComment/top-level/inline",
			input: "text <!-- comment --> text",
			want: []wantBlock{
				text("", "text "),
				comment("", "<!-- comment -->"),
				text("", " text"),
			},
		},
		{
			name:  "HTMLComment/top-level/inline-multiple",
			input: "text <!-- comment --><!--\ncomment --> text",
			want: []wantBlock{
				text("", "text "),
				comment("", "<!-- comment -->"),
				comment("", "<!--\ncomment -->"),
				text("", " text"),
			},
		},
		{
			name:  "HTMLComment/top-level/inline-at-line-end",
			input: "text <!-- comment -->",
			want: []wantBlock{
				text("", "text "),
				comment("", "<!-- comment -->"),
			},
		},
		{
			name: "HTMLComment/top-level/block",
			input: joinLines(
				"text",
				"<!--",
				"comment",
				"-->",
				"text",
			),
			want: []wantBlock{
				text("", "text\n"),
				comment("", "<!--\ncomment\n-->\n"),
				text("", "text"),
			},
		},
		{
			name:  "HTMLComment/top-level/block-empty",
			input: "<!---->\n",
			want: []wantBlock{
				comment("", "<!---->\n"),
			},
		},
		{
			name: "HTMLComment/top-level/block-consecutive",
			input: joinLines(
				"<!-- a -->",
				"",
				"<!-- b -->",
				"",
			),
			want: []wantBlock{
				comment("", "<!-- a -->\n"),
				text("", "\n"),
				comment("", "<!-- b -->\n"),
			},
		},
		{
			name:  "HTMLComment/top-level/in-heading-line",
			input: "# Title <!-- comment -->\n",
			want: []wantBlock{
				text("", "# Title "),
				comment("", "<!-- comment -->"),
				text("", "\n"),
			},
		},

		// HTMLComment — inside containers.
		{
			name:  "HTMLComment/blockquote/inline-multiple",
			input: "> text <!-- comment --><!--\n> comment --> text",
			want: []wantBlock{
				text("> ", "text "),
				comment("", "<!-- comment -->"),
				comment("", "<!--\ncomment -->"),
				text("> ", " text"),
			},
		},
		{
			name: "HTMLComment/blockquote/nested/block",
			input: joinLines(
				"> text",
				"> > <!--",
				"> > comment",
				"> > -->",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				comment("> > ", "<!--\ncomment\n-->\n"),
				text("> ", "text"),
			},
		},
		{
			name:  "HTMLComment/list/dash/inline-multiple",
			input: "- text <!-- comment --><!--\n  comment --> text",
			want: []wantBlock{
				text("- ", "text "),
				comment("", "<!-- comment -->"),
				comment("", "<!--\ncomment -->"),
				text("- ", " text"),
			},
		},
		{
			name:  "HTMLComment/list/dash/nested/inline-multiple",
			input: "  - text <!-- comment --><!--\n    comment --> text",
			want: []wantBlock{
				text("  - ", "text "),
				comment("", "<!-- comment -->"),
				comment("", "<!--\ncomment -->"),
				text("  - ", " text"),
			},
		},
		{
			name: "HTMLComment/list/dash/block",
			input: joinLines(
				"- text",
				"- <!--",
				"  comment",
				"  -->",
				"- text",
			),
			want: []wantBlock{
				text("- ", "text\n"),
				comment("- ", "<!--\ncomment\n-->\n"),
				text("- ", "text"),
			},
		},
		{
			name: "HTMLComment/list/dash/nested/block",
			input: joinLines(
				"- text",
				"  - <!--",
				"    comment",
				"    -->",
				"- text",
			),
			want: []wantBlock{
				text("- ", "text\n"),
				comment("  - ", "<!--\ncomment\n-->\n"),
				text("- ", "text"),
			},
		},
		{
			name: "HTMLComment/list/star/block",
			input: joinLines(
				"* text",
				"* <!--",
				"  comment",
				"  -->",
				"* text",
			),
			want: []wantBlock{
				text("* ", "text\n"),
				comment("* ", "<!--\ncomment\n-->\n"),
				text("* ", "text"),
			},
		},
		{
			name: "HTMLComment/list/ordered/block",
			input: joinLines(
				"1. text",
				"2. <!--",
				"   comment",
				"   -->",
				"3. text",
			),
			want: []wantBlock{
				text("1. ", "text\n"),
				comment("2. ", "<!--\ncomment\n-->\n"),
				text("3. ", "text"),
			},
		},
		{
			name:  "HTMLComment/blockquote/list/inline-multiple",
			input: "> - text <!-- comment --><!--\n>   comment --> text",
			want: []wantBlock{
				text("> - ", "text "),
				comment("", "<!-- comment -->"),
				comment("", "<!--\ncomment -->"),
				text("> - ", " text"),
			},
		},
		{
			name: "HTMLComment/blockquote/list/block",
			input: joinLines(
				"> text",
				"> - <!--",
				">   comment",
				">   -->",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				comment("> - ", "<!--\ncomment\n-->\n"),
				text("> ", "text"),
			},
		},
		{
			name: "HTMLComment/blockquote/nested/list/block",
			input: joinLines(
				"> text",
				"> > - <!--",
				"> >   comment",
				"> >   -->",
				"> text",
			),
			want: []wantBlock{
				text("> ", "text\n"),
				comment("> > - ", "<!--\ncomment\n-->\n"),
				text("> ", "text"),
			},
		},
		{
			name: "HTMLComment/list/ordered/blockquote/block",
			input: joinLines(
				"1. text",
				"   > <!--",
				"   > comment",
				"   > -->",
				"2. text",
			),
			want: []wantBlock{
				text("1. ", "text\n"),
				comment("   > ", "<!--\ncomment\n-->\n"),
				text("2. ", "text"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			blocks, err := internal.MakeBlocksFromMarkdown([]byte(tt.input))

			// then
			require.NoError(t, err)
			require.Len(
				t,
				blocks,
				len(tt.want),
				"expected %d blocks, got %d",
				len(tt.want),
				len(blocks),
			)
			for i, w := range tt.want {
				assert.Equal(t, w.kind, blocks[i].Kind(), "block[%d] kind", i)
				assert.Equal(t, w.indent, string(blocks[i].Indent()), "block[%d] indent", i)
				assert.Equal(t, w.content, string(blocks[i].Content()), "block[%d] content", i)
			}
		})
	}
}
