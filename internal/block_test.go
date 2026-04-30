package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeBlockFromRaw(t *testing.T) {
	// given
	kind := internal.BlockKindFencedCode
	content := "```bash\necho hello\n```\n"
	indent := "> "
	continuation := true

	// when
	got := internal.MakeBlockFromRaw(kind, content, indent, continuation)

	// then
	assert.Equal(t, kind, got.Kind())
	assert.Equal(t, content, got.Content())
	assert.Equal(t, indent, got.Indent())
	assert.Equal(t, continuation, got.Continuation())
}

func TestMakeBlocksFromMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []internal.Block
	}{
		// Text — baseline cases that should produce no FencedCode or HTMLComment blocks.
		{
			name:  "Text/heading",
			input: "# Hello\n",
			want: []internal.Block{
				text("", "# Hello\n", false),
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
			want: []internal.Block{
				text("", "# Level 1\n", false),
				text("", "\n", false),
				text("", "some text\n", false),
				text("", "\n", false),
				text("", "## Level 2\n", false),
				text("", "\n", false),
				text("", "more text\n", false),
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
			want: []internal.Block{
				text("", "text\nand more text\n", false),
				text("", "\n", false),
				text("", "last line", false),
			},
		},
		{
			name: "Text/paragraph/line-break",
			input: joinLines(
				"text",
				"and more text  ",
				"after line break",
			),
			want: []internal.Block{
				text("", "text\nand more text  \nafter line break", false),
			},
		},
		{
			name: "Text/indented-code-as-text",
			input: joinLines(
				"    echo \"hello, \"",
				"    echo \"world\"",
			),
			want: []internal.Block{
				text("", "    echo \"hello, \"\n    echo \"world\"", false),
			},
		},
		{
			name:  "Text/inline-code-span",
			input: "text `echo \"hello, \" text`",
			want: []internal.Block{
				text("", "text `echo \"hello, \" text`", false),
			},
		},
		{
			name:  "Text/html-block-not-comment",
			input: "<div>\nhello\n</div>\n",
			want: []internal.Block{
				text("", "<div>\nhello\n</div>\n", false),
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
			want: []internal.Block{
				text("> ", "hello, \n", false),
				text("> ", "world\n", false),
				text("> ", "\n", false),
				text("> ", "again", false),
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
			want: []internal.Block{
				code("", "```bash\necho \"hello\"\n```\n", false),
			},
		},
		{
			name: "FencedCode/top-level/backtick-no-trailing-newline",
			input: joinLines(
				"```bash",
				"echo \"hello\"",
				"```",
			),
			want: []internal.Block{
				code("", "```bash\necho \"hello\"\n```", false),
			},
		},
		{
			name: "FencedCode/top-level/tilde",
			input: joinLines(
				"~~~bash",
				"echo \"hello\"",
				"~~~",
			),
			want: []internal.Block{
				code("", "~~~bash\necho \"hello\"\n~~~", false),
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
			want: []internal.Block{
				code("", "````md\n```bash\necho \"hello\"\n```\n````", false),
			},
		},
		{
			name: "FencedCode/top-level/no-info-string",
			input: joinLines(
				"```",
				"echo \"hello\"",
				"```",
			),
			want: []internal.Block{
				code("", "```\necho \"hello\"\n```", false),
			},
		},
		{
			name: "FencedCode/top-level/empty",
			input: joinLines(
				"```bash",
				"```",
			),
			want: []internal.Block{
				code("", "```bash\n```", false),
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
			want: []internal.Block{
				code("", "```bash\necho a\n```\n", false),
				text("", "\n", false),
				code("", "```bash\necho b\n```", false),
			},
		},
		{
			name: "FencedCode/top-level/indented-1-space",
			input: joinLines(
				" ```bash",
				" echo \"hello\"",
				" ```",
			),
			want: []internal.Block{
				code(" ", "```bash\necho \"hello\"\n```", false),
			},
		},
		{
			name: "FencedCode/top-level/indented-3-spaces",
			input: joinLines(
				"   ```bash",
				"   echo \"hello\"",
				"   ```",
			),
			want: []internal.Block{
				code("   ", "```bash\necho \"hello\"\n```", false),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				code("> ", "```bash\necho \"hello\"\n```\n", false),
				text("> ", "text", false),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				code("> > ", "```bash\necho \"hello\"\n```\n", false),
				text("> ", "text", false),
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
			want: []internal.Block{
				text("- ", "text\n", false),
				code("- ", "```bash\necho \"hello\"\n```\n", false),
				text("- ", "text", false),
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
			want: []internal.Block{
				text("- ", "text\n", false),
				code("  - ", "```bash\necho \"hello\"\n```\n", false),
				text("- ", "text", false),
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
			want: []internal.Block{
				text("+ ", "text\n", false),
				code("+ ", "~~~bash\necho \"hello\"\n~~~\n", false),
				text("+ ", "text", false),
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
			want: []internal.Block{
				text("1. ", "text\n", false),
				code("2. ", "```bash\necho \"hello\"\n```\n", false),
				text("3. ", "text", false),
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
			want: []internal.Block{
				text("1. ", "text\n", false),
				code("   1. ", "```bash\necho \"hello\"\n```\n", false),
				text("2. ", "text", false),
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
			want: []internal.Block{
				text("- ", "text\n", false),
				code("  > ", "```bash\necho \"hello\"\n```\n", false),
				text("- ", "text", false),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				text("> - ", "item\n", false),
				code(">   - ", "```bash\necho \"hello\"\n```\n", false),
				text("> ", "text", false),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				code("> 1. ", "```bash\necho \"hello\"\n```\n", false),
				text("> ", "text", false),
			},
		},

		// HTMLComment — top-level (CommonMark §4.6 type 2).
		{
			name:  "HTMLComment/top-level/inline",
			input: "text <!-- comment --> text",
			want: []internal.Block{
				text("", "text ", false),
				cmnt("", "<!-- comment -->", true),
				text("", " text", true),
			},
		},
		{
			name:  "HTMLComment/top-level/inline-multiple",
			input: "text <!-- comment --><!--\ncomment --> text",
			want: []internal.Block{
				text("", "text ", false),
				cmnt("", "<!-- comment -->", true),
				cmnt("", "<!--\ncomment -->", true),
				text("", " text", true),
			},
		},
		{
			name:  "HTMLComment/top-level/inline-at-line-end",
			input: "text <!-- comment -->",
			want: []internal.Block{
				text("", "text ", false),
				cmnt("", "<!-- comment -->", true),
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
			want: []internal.Block{
				text("", "text\n", false),
				cmnt("", "<!--\ncomment\n-->\n", false),
				text("", "text", false),
			},
		},
		{
			name:  "HTMLComment/top-level/block-empty",
			input: "<!---->\n",
			want: []internal.Block{
				cmnt("", "<!---->\n", false),
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
			want: []internal.Block{
				cmnt("", "<!-- a -->\n", false),
				text("", "\n", false),
				cmnt("", "<!-- b -->\n", false),
			},
		},
		{
			name:  "HTMLComment/top-level/in-heading-line",
			input: "# Title <!-- comment -->\n",
			want: []internal.Block{
				text("", "# Title ", false),
				cmnt("", "<!-- comment -->", true),
				text("", "\n", true),
			},
		},

		// HTMLComment — inside containers.
		{
			name:  "HTMLComment/blockquote/inline-multiple",
			input: "> text <!-- comment --><!--\n> comment --> text",
			want: []internal.Block{
				text("> ", "text ", false),
				cmnt("> ", "<!-- comment -->", true),
				cmnt("> ", "<!--\ncomment -->", true),
				text("> ", " text", true),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				cmnt("> > ", "<!--\ncomment\n-->\n", false),
				text("> ", "text", false),
			},
		},
		{
			name:  "HTMLComment/list/dash/inline-multiple",
			input: "- text <!-- comment --><!--\n  comment --> text",
			want: []internal.Block{
				text("- ", "text ", false),
				cmnt("- ", "<!-- comment -->", true),
				cmnt("- ", "<!--\ncomment -->", true),
				text("- ", " text", true),
			},
		},
		{
			name:  "HTMLComment/list/dash/nested/inline-multiple",
			input: "  - text <!-- comment --><!--\n    comment --> text",
			want: []internal.Block{
				text("  - ", "text ", false),
				cmnt("  - ", "<!-- comment -->", true),
				cmnt("  - ", "<!--\ncomment -->", true),
				text("  - ", " text", true),
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
			want: []internal.Block{
				text("- ", "text\n", false),
				cmnt("- ", "<!--\ncomment\n-->\n", false),
				text("- ", "text", false),
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
			want: []internal.Block{
				text("- ", "text\n", false),
				cmnt("  - ", "<!--\ncomment\n-->\n", false),
				text("- ", "text", false),
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
			want: []internal.Block{
				text("* ", "text\n", false),
				cmnt("* ", "<!--\ncomment\n-->\n", false),
				text("* ", "text", false),
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
			want: []internal.Block{
				text("1. ", "text\n", false),
				cmnt("2. ", "<!--\ncomment\n-->\n", false),
				text("3. ", "text", false),
			},
		},
		{
			name:  "HTMLComment/blockquote/list/inline-multiple",
			input: "> - text <!-- comment --><!--\n>   comment --> text",
			want: []internal.Block{
				text("> - ", "text ", false),
				cmnt("> - ", "<!-- comment -->", true),
				cmnt("> - ", "<!--\ncomment -->", true),
				text("> - ", " text", true),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				cmnt("> - ", "<!--\ncomment\n-->\n", false),
				text("> ", "text", false),
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
			want: []internal.Block{
				text("> ", "text\n", false),
				cmnt("> > - ", "<!--\ncomment\n-->\n", false),
				text("> ", "text", false),
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
			want: []internal.Block{
				text("1. ", "text\n", false),
				cmnt("   > ", "<!--\ncomment\n-->\n", false),
				text("2. ", "text", false),
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
				assert.Equal(t, w.Kind(), blocks[i].Kind(), "block[%d] kind", i)
				assert.Equal(t, w.Indent(), blocks[i].Indent(), "block[%d] indent", i)
				assert.Equal(t, w.Content(), blocks[i].Content(), "block[%d] content", i)
				assert.Equal(
					t,
					w.Continuation(),
					blocks[i].Continuation(),
					"block[%d] continuation",
					i,
				)
			}
		})
	}

	invalidTests := []struct {
		name        string
		input       []byte
		wantErrText string
	}{
		{
			name:        "invalid-utf8",
			input:       []byte{'#', ' ', 0xe9, '\n'},
			wantErrText: "not valid UTF-8",
		},
		{
			name: "FencedCode/top-level/unclosed-runs-to-eof",
			input: []byte(joinLines(
				"```bash",
				"echo \"hello\"",
			)),
			wantErrText: "unclosed fenced code block",
		},
		{
			name:        "FencedCode/top-level/opening-line-runs-to-eof",
			input:       []byte("```bash\n"),
			wantErrText: "unclosed fenced code block",
		},
		{
			name: "FencedCode/top-level/wrong-closing-fence-char",
			input: []byte(joinLines(
				"```bash",
				"echo \"hello\"",
				"~~~",
			)),
			wantErrText: "unclosed fenced code block",
		},
		{
			name: "FencedCode/top-level/short-closing-fence",
			input: []byte(joinLines(
				"````bash",
				"echo \"hello\"",
				"```",
			)),
			wantErrText: "unclosed fenced code block",
		},
		{
			name: "FencedCode/top-level/closing-fence-with-trailing-text",
			input: []byte(joinLines(
				"```bash",
				"echo \"hello\"",
				"``` nope",
			)),
			wantErrText: "unclosed fenced code block",
		},
		{
			name: "FencedCode/blockquote/unclosed-runs-to-eof",
			input: []byte(joinLines(
				"> ```bash",
				"> echo \"hello\"",
			)),
			wantErrText: "unclosed fenced code block",
		},
		{
			name: "HTMLComment/top-level/unclosed-block-runs-to-eof",
			input: []byte(joinLines(
				"<!--",
				"comment",
			)),
			wantErrText: "unclosed HTML comment",
		},
		{
			name:        "HTMLComment/top-level/unclosed-single-line-runs-to-eof",
			input:       []byte("<!-- comment"),
			wantErrText: "unclosed HTML comment",
		},
		{
			name: "HTMLComment/blockquote/unclosed-runs-to-eof",
			input: []byte(joinLines(
				"> <!--",
				"> comment",
			)),
			wantErrText: "unclosed HTML comment",
		},
		{
			name: "HTMLComment/list/unclosed-runs-to-eof",
			input: []byte(joinLines(
				"- <!--",
				"  comment",
			)),
			wantErrText: "unclosed HTML comment",
		},
	}

	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			_, err := internal.MakeBlocksFromMarkdown(tt.input)

			// then
			require.ErrorContains(t, err, tt.wantErrText)
		})
	}
}
