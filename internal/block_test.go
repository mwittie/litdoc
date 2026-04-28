package internal_test

import (
	"strings"
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeBlockFromRaw(t *testing.T) {
	// given
	kind := internal.BlockKindFencedCode
	content := []byte("```bash\necho hello\n```\n")

	// when
	got := internal.MakeBlockFromRaw(kind, content)

	// then
	assert.Equal(t, kind, got.Kind())
	assert.Equal(t, content, got.Content())
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestMakeBlocksFromMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []struct {
			kind    internal.BlockKind
			content string
		}
	}{
		{
			name:  "heading",
			input: "# Hello\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "# Hello\n"},
			},
		},
		{
			name: "nested heading and text",
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
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "# Level 1\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindText, "some text\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindText, "## Level 2\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindText, "more text\n"},
			},
		},
		{
			name: "paragraphs grouped together",
			input: joinLines(
				"text",
				"and more text",
				"",
				"last line",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text\nand more text\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindText, "last line"},
			},
		},
		{
			name: "line break grouped with paragraph",
			input: joinLines(
				"text",
				"and more text  ",
				"after line break",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text\nand more text  \nafter line break"},
			},
		},
		{
			name: "fenced code block as single block",
			input: joinLines(
				"```bash",
				"echo \"hello\"",
				"```",
				"",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindFencedCode, "```bash\necho \"hello\"\n```\n"},
			},
		},
		{
			name: "fenced code block with no trailing newline",
			input: joinLines(
				"```bash",
				"echo \"hello\"",
				"```",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindFencedCode, "```bash\necho \"hello\"\n```"},
			},
		},
		{
			name: "tilde fenced code block as single block",
			input: joinLines(
				"~~~bash",
				"echo \"hello\"",
				"~~~",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindFencedCode, "~~~bash\necho \"hello\"\n~~~"},
			},
		},
		{
			name: "longer backtick fence as single block",
			input: joinLines(
				"````md",
				"```bash",
				"echo \"hello\"",
				"```",
				"````",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{
					internal.BlockKindFencedCode,
					"````md\n```bash\necho \"hello\"\n```\n````",
				},
			},
		},
		{
			name: "indented code block",
			input: joinLines(
				"    echo \"hello, \"",
				"    echo \"world\"",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{
					internal.BlockKindText,
					"    echo \"hello, \"\n    echo \"world\"",
				},
			},
		},
		{
			name:  "inline code block",
			input: "text `echo \"hello, \" text`",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text `echo \"hello, \" text`"},
			},
		},
		{
			name:  "html comment inline comment",
			input: "text <!-- comment --> text",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text "},
				{internal.BlockKindHTMLComment, "<!-- comment -->"},
				{internal.BlockKindText, " text"},
			},
		},
		{
			name:  "multiple inline html comments",
			input: "text <!-- comment --><!--\ncomment --> text",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text "},
				{internal.BlockKindHTMLComment, "<!-- comment -->"},
				{internal.BlockKindHTMLComment, "<!--\ncomment -->"},
				{internal.BlockKindText, " text"},
			},
		},
		{
			name: "html comment block comment",
			input: joinLines(
				"text",
				"<!--",
				"comment",
				"-->",
				"text",
			),
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text\n"},
				{internal.BlockKindHTMLComment, "<!--\ncomment\n-->\n"},
				{internal.BlockKindText, "text"},
			},
		},
		{
			name:  "html block is not a comment",
			input: "<div>\nhello\n</div>\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "<div>\nhello\n</div>\n"},
			},
		},
		{
			name:  "mixed content",
			input: "# Title\n\n```go\nfmt.Println()\n```\n\n<!--\nnote\n-->\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "# Title\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindFencedCode, "```go\nfmt.Println()\n```\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindHTMLComment, "<!--\nnote\n-->\n"},
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
				assert.Equal(
					t,
					w.kind,
					blocks[i].Kind(),
					"block[%d] kind",
					i,
				)
				assert.Equal(
					t,
					w.content,
					string(blocks[i].Content()),
					"block[%d] content",
					i,
				)
			}
		})
	}
}
