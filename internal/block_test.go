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
	content := []byte("```bash\necho hello\n```\n")

	// when
	got := internal.MakeBlockFromRaw(kind, content)

	// then
	assert.Equal(t, kind, got.Kind())
	assert.Equal(t, content, got.Content())
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
			name:  "text",
			input: "text\nand more text",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "text\nand more text"},
			},
		},
		{
			name:  "fenced code block",
			input: "```bash\necho \"hello\"\n```\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindFencedCode, "```bash\necho \"hello\"\n```\n"},
			},
		},
		{
			name:  "fenced code block in nested list keeps indent in content",
			input: "- Level 1\n\n  - Level 2\n\n    ```bash | litdoc\n    echo \"hello\"\n    ```\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "- "},
				{internal.BlockKindText, "Level 1\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindText, "  "},
				{internal.BlockKindText, "- "},
				{internal.BlockKindText, "Level 2\n"},
				{internal.BlockKindText, "\n"},
				{
					internal.BlockKindFencedCode,
					"    ```bash | litdoc\n    echo \"hello\"\n    ```\n",
				},
			},
		},
		{
			name:  "tilde fenced code block",
			input: "~~~bash\necho \"hello\"\n~~~\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindFencedCode, "~~~bash\necho \"hello\"\n~~~\n"},
			},
		},
		{
			name:  "longer backtick fence",
			input: "````bash\necho \"hello\"\n````\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindFencedCode, "````bash\necho \"hello\"\n````\n"},
			},
		},
		{
			name:  "quad fenced verbatim block containing litdoc source",
			input: "````md\n```bash | litdoc\necho \"hello\"\n```\n````",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{
					internal.BlockKindFencedCode,
					"````md\n```bash | litdoc\necho \"hello\"\n```\n````",
				},
			},
		},
		{
			name:  "indented code block",
			input: "    echo \"hello\"\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "    echo \"hello\"\n"},
			},
		},
		{
			name:  "html comment block",
			input: "<!--\ncomment\n-->\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindHTMLComment, "<!--\ncomment\n-->\n"},
			},
		},
		{
			name:  "html comment block in nested list keeps indent in content",
			input: "- Level 1\n\n  - Level 2\n\n    <!--\n    comment\n    -->\n    text\n",
			want: []struct {
				kind    internal.BlockKind
				content string
			}{
				{internal.BlockKindText, "- "},
				{internal.BlockKindText, "Level 1\n"},
				{internal.BlockKindText, "\n"},
				{internal.BlockKindText, "  "},
				{internal.BlockKindText, "- "},
				{internal.BlockKindText, "Level 2\n"},
				{internal.BlockKindText, "\n"},
				{
					internal.BlockKindHTMLComment,
					"    <!--\n    comment\n    -->\n",
				},
				{internal.BlockKindText, "    text\n"},
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
			require.Len(t, blocks, len(tt.want))
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
