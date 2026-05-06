package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
)

func TestInfoStringFromBlock(t *testing.T) {
	tests := []struct {
		name  string
		block internal.Block
		want  internal.InfoString
	}{
		{
			name:  "text block",
			block: text("", "hello", false),
			want:  internal.InfoString{},
		},
		{
			name: "fenced code/backtick/without-litdoc",
			block: code(
				"",
				joinLines(
					"```bash",
					"echo hello",
					"```",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "bash"},
		},
		{
			name: "fenced code/backtick/with-litdoc",
			block: code(
				"",
				joinLines(
					"```bash | litdoc",
					"echo hello",
					"```",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "bash", Litdoc: true},
		},
		{
			name: "fenced code/tilde/with-litdoc",
			block: code(
				"",
				joinLines(
					"~~~sh | litdoc",
					"echo hello",
					"~~~",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "sh", Litdoc: true},
		},
		{
			name: "fenced code/no-info-string",
			block: code(
				"",
				joinLines(
					"```",
					"echo hello",
					"```",
					"",
				),
				false,
			),
			want: internal.InfoString{},
		},
		{
			name: "fenced code/trims-language",
			block: code(
				"",
				joinLines(
					"```  bash  | litdoc",
					"echo hello",
					"```",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "bash", Litdoc: true},
		},
		{
			name: "fenced code/litdoc-prefix",
			block: code(
				"",
				joinLines(
					"```bash | litdoc-output",
					"echo hello",
					"```",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "bash", Litdoc: true},
		},
		{
			name: "html comment/without-litdoc",
			block: cmnt(
				"",
				joinLines(
					"<!-- bash",
					"echo hello",
					"-->",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "bash"},
		},
		{
			name: "html comment/with-litdoc",
			block: cmnt(
				"",
				joinLines(
					"<!-- bash | litdoc",
					"echo hello",
					"-->",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "bash", Litdoc: true},
		},
		{
			name: "html comment/unsupported-litdoc-language",
			block: cmnt(
				"",
				joinLines(
					"<!-- go | litdoc",
					"fmt.Println()",
					"-->",
					"",
				),
				false,
			),
			want: internal.InfoString{Lang: "go", Litdoc: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, internal.InfoStringFromBlock(tt.block))
		})
	}
}

func TestCodeBody(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "single line",
			content: joinLines(
				"```bash",
				"echo hello",
				"```",
				"",
			),
			want: "echo hello\n",
		},
		{
			name: "multi-line body",
			content: joinLines(
				"```bash",
				"line1",
				"line2",
				"line3",
				"```",
				"",
			),
			want: "line1\nline2\nline3\n",
		},
		{
			name: "no info string",
			content: joinLines(
				"```",
				"echo hello",
				"```",
				"",
			),
			want: "echo hello\n",
		},
		{
			name: "longer fence",
			content: joinLines(
				"````bash",
				"echo hello",
				"````",
				"",
			),
			want: "echo hello\n",
		},
		{
			name: "no trailing newline on content",
			content: joinLines(
				"```bash",
				"echo hello",
				"```",
			),
			want: "echo hello\n",
		},
		{
			name: "body line containing backticks",
			content: joinLines(
				"```bash",
				"echo 'hello ```'",
				"```",
				"",
			),
			want: "echo 'hello ```'\n",
		},
		{
			name: "empty body",
			content: joinLines(
				"```bash",
				"```",
				"",
			),
			want: "",
		},
		{
			name: "missing closing fence",
			content: joinLines(
				"```bash",
				"echo hello",
				"",
			),
			want: "",
		},
		{
			name:    "not a fenced block",
			content: "echo hello",
			want:    "",
		},
		{
			name: "tilde fence",
			content: joinLines(
				"~~~bash",
				"echo hello",
				"~~~",
				"",
			),
			want: "echo hello\n",
		},
		{
			name: "tilde fence with info string",
			content: joinLines(
				"~~~bash | litdoc",
				"echo hello",
				"~~~",
				"",
			),
			want: "echo hello\n",
		},
		{
			name: "html comment",
			content: joinLines(
				"<!--bash | litdoc",
				"echo hello",
				"-->",
				"",
			),
			want: "echo hello\n",
		},
		{
			name: "html comment multi-line",
			content: joinLines(
				"<!--bash | litdoc",
				"line1",
				"line2",
				"-->",
				"",
			),
			want: "line1\nline2\n",
		},
		{
			name: "html comment no trailing newline",
			content: joinLines(
				"<!--bash | litdoc",
				"echo hello",
				"-->",
			),
			want: "echo hello\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, internal.CodeBody(tt.content))
		})
	}
}
