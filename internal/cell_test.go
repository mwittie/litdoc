package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticCell(t *testing.T) {
	t.Run("renders raw content", func(t *testing.T) {
		// given
		content := "hello"
		cell := internal.MakeStaticCellFromRaw(content)

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, content, gotContent)
	})

	t.Run("executes to itself", func(t *testing.T) {
		// given
		cell := internal.MakeStaticCellFromRaw("hello")

		// when
		gotCell, err := cell.Execute()

		// then
		require.NoError(t, err)
		assert.Equal(t, cell, gotCell)
	})
}

func TestParseInfoString(t *testing.T) {
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
			got := internal.ParseInfoString(tt.block)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBashCell(t *testing.T) {
	t.Run("without output", func(t *testing.T) {
		// given
		code := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		cell := internal.MakeBashCellFromRaw(code, "")

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, code, gotContent)
	})

	t.Run("with output", func(t *testing.T) {
		// given
		fencedCode := joinLines(
			"```bash",
			"echo hello",
			"```",
		)
		output := "hello"
		cell := internal.MakeBashCellFromRaw(fencedCode, output)

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, fencedCode+"\n"+output, gotContent)
	})

	t.Run("executes to itself", func(t *testing.T) {
		// given
		fencedCode := joinLines(
			"```bash",
			"echo hello",
			"```",
			"",
		)
		cell := internal.MakeBashCellFromRaw(fencedCode, "")

		// when
		gotCell, err := cell.Execute()

		// then
		require.NoError(t, err)
		rendered, err := gotCell.Render()
		require.NoError(t, err)
		assert.Equal(t, fencedCode, rendered)
	})
}

func TestExecute(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		result := NewMockCell(t)
		cell := NewMockCell(t)
		cell.EXPECT().Execute().Return(result, nil)
		cells := []internal.Cell{cell}

		// when
		gotCells, err := internal.Execute(cells)

		// then
		require.NoError(t, err)
		require.Len(t, gotCells, 1)
		assert.Equal(t, result, gotCells[0])
	})

	t.Run("cell.Execute fails", func(t *testing.T) {
		// given
		cell := NewMockCell(t)
		cell.EXPECT().Execute().Return(nil, assert.AnError)
		cells := []internal.Cell{cell}

		// when
		_, err := internal.Execute(cells)

		// then
		require.ErrorContains(t, err, "executing cell")
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestCompose(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		cell1 := NewMockCell(t)
		cell1.EXPECT().Render().Return("hello", nil)
		cell2 := NewMockCell(t)
		cell2.EXPECT().Render().Return(" world", nil)
		cells := []internal.Cell{cell1, cell2}

		// when
		got, err := internal.Compose(cells)

		// then
		require.NoError(t, err)
		assert.Equal(t, "hello world", got)
	})

	t.Run("cell.Render fails", func(t *testing.T) {
		// given
		cell := NewMockCell(t)
		cell.EXPECT().Render().Return("", assert.AnError)
		cells := []internal.Cell{cell}

		// when
		_, err := internal.Compose(cells)

		// then
		require.ErrorContains(t, err, "rendering cell")
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestClassify(t *testing.T) {
	type wantCell struct {
		kind     string
		rendered string
	}

	tests := []struct {
		name        string
		blocks      []internal.Block
		want        []wantCell
		wantErrText string
	}{
		{
			name: "Text/top-level",
			blocks: []internal.Block{
				text("", "hello", false),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "hello"},
			},
		},
		{
			name: "Text/blockquote/multiline",
			blocks: []internal.Block{
				text("> ", "hello, \nworld\n", false),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "> hello, \n> world\n"},
			},
		},
		{
			name: "Text/list/dash",
			blocks: []internal.Block{
				text("- ", "text\ncontinued", false),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "- text\n  continued"},
			},
		},
		{
			name: "Text/list/ordered",
			blocks: []internal.Block{
				text("2. ", "text\ncontinued", false),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "2. text\n   continued"},
			},
		},
		{
			name: "Text/blockquote/list",
			blocks: []internal.Block{
				text("> - ", "text\ncontinued", false),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "> - text\n>   continued"},
			},
		},
		{
			name: "FencedCode/non-litdoc/top-level",
			blocks: []internal.Block{
				code("", joinLines(
					"```bash",
					"echo hello",
					"```",
					"",
				), false),
			},
			want: []wantCell{
				{
					kind: "StaticCell", rendered: joinLines(
						"```bash",
						"echo hello",
						"```",
						"",
					),
				},
			},
		},
		{
			name: "FencedCode/non-litdoc/list",
			blocks: []internal.Block{
				code("- ", joinLines(
					"```bash",
					"echo hello",
					"```",
					"",
				), false),
			},
			want: []wantCell{
				{
					kind: "StaticCell", rendered: joinLines(
						"- ```bash",
						"  echo hello",
						"  ```",
						"",
					),
				},
			},
		},
		{
			name: "FencedCode/litdoc/bash",
			blocks: []internal.Block{
				code("", joinLines(
					"```bash | litdoc",
					"echo hello",
					"```",
					"",
				), false),
			},
			want: []wantCell{
				{
					kind: "BashCell", rendered: joinLines(
						"```bash | litdoc",
						"echo hello",
						"```",
						"",
					),
				},
			},
		},
		{
			name: "FencedCode/litdoc/unsupported-language",
			blocks: []internal.Block{
				code("", joinLines(
					"```go | litdoc",
					"fmt.Println()",
					"```",
					"",
				), false),
			},
			wantErrText: "unsupported language",
		},
		{
			name: "HTMLComment/block/list",
			blocks: []internal.Block{
				cmnt("- ", joinLines(
					"<!--",
					"comment",
					"-->",
					"",
				), false),
			},
			want: []wantCell{
				{
					kind: "StaticCell", rendered: joinLines(
						"- <!--",
						"  comment",
						"  -->",
						"",
					),
				},
			},
		},
		{
			name: "HTMLComment/litdoc/bash",
			blocks: []internal.Block{
				cmnt("", joinLines(
					"<!--bash | litdoc",
					"echo hello",
					"-->",
					"",
				), false),
			},
			want: []wantCell{
				{
					kind: "BashCell", rendered: joinLines(
						"<!--bash | litdoc",
						"echo hello",
						"-->",
						"",
					),
				},
			},
		},
		{
			name: "HTMLComment/inline-continuation/list",
			blocks: []internal.Block{
				text("- ", "text ", false),
				cmnt("- ", "<!-- comment -->", true),
				cmnt("- ", "<!--\ncomment -->", true),
				text("- ", " text", true),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "- text "},
				{kind: "StaticCell", rendered: "<!-- comment -->"},
				{kind: "StaticCell", rendered: "<!--\n  comment -->"},
				{kind: "StaticCell", rendered: " text"},
			},
		},
		{
			name: "HTMLComment/inline-continuation/blockquote-list",
			blocks: []internal.Block{
				text("> - ", "text ", false),
				cmnt("> - ", "<!-- comment -->", true),
				cmnt("> - ", "<!--\ncomment -->", true),
				text("> - ", " text", true),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "> - text "},
				{kind: "StaticCell", rendered: "<!-- comment -->"},
				{kind: "StaticCell", rendered: "<!--\n>   comment -->"},
				{kind: "StaticCell", rendered: " text"},
			},
		},
		{
			name: "Mixed/independent-blocks",
			blocks: []internal.Block{
				text("", "before", false),
				code("", joinLines(
					"```bash | litdoc",
					"echo hello",
					"```",
					"",
				), false),
				text("", "after", false),
			},
			want: []wantCell{
				{kind: "StaticCell", rendered: "before"},
				{
					kind: "BashCell", rendered: joinLines(
						"```bash | litdoc",
						"echo hello",
						"```",
						"",
					),
				},
				{kind: "StaticCell", rendered: "after"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// when
			cells, err := internal.Classify(tt.blocks)

			// then
			if tt.wantErrText != "" {
				require.ErrorContains(t, err, tt.wantErrText)
				return
			}

			require.NoError(t, err)
			require.Len(t, cells, len(tt.want))
			wantComposed := ""
			for i, w := range tt.want {
				assert.Equal(t, w.kind, cellKind(cells[i]), "cell[%d] kind", i)
				rendered, err := cells[i].Render()
				require.NoError(t, err, "cell[%d] render", i)
				assert.Equal(t, w.rendered, rendered, "cell[%d] rendered", i)
				wantComposed += w.rendered
			}

			gotComposed, err := internal.Compose(cells)
			require.NoError(t, err)
			assert.Equal(t, wantComposed, gotComposed)
		})
	}
}

func cellKind(cell internal.Cell) string {
	switch cell.(type) {
	case internal.StaticCell:
		return "StaticCell"
	case internal.BashCell:
		return "BashCell"
	default:
		return "unknown"
	}
}
