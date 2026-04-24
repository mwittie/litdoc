package internal_test

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeStaticCellFromRaw(t *testing.T) {
	// given
	content := "hello"

	// when
	gotCell := internal.MakeStaticCellFromRaw(content)

	// then
	got, err := gotCell.Render()
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestStaticCellExecute(t *testing.T) {
	// given
	cell := internal.MakeStaticCellFromRaw("hello")

	// when
	gotCell, err := cell.Execute()

	// then
	require.NoError(t, err)
	assert.Equal(t, cell, gotCell)
}

func TestStaticCellRender(t *testing.T) {
	// given
	content := "hello"
	cell := internal.MakeStaticCellFromRaw(content)

	// when
	gotContent, err := cell.Render()

	// then
	require.NoError(t, err)
	assert.Equal(t, content, gotContent)
}

func TestParseInfoString(t *testing.T) {
	tests := []struct {
		name  string
		block internal.Block
		want  internal.InfoString
	}{
		{
			name: "text block",
			block: internal.MakeBlockFromRaw(
				internal.BlockKindText,
				[]byte("hello"),
			),
			want: internal.InfoString{},
		},
		{
			name: "fenced code without litdoc",
			block: internal.MakeBlockFromRaw(
				internal.BlockKindFencedCode,
				[]byte("```bash\necho hello\n```\n"),
			),
			want: internal.InfoString{Lang: "bash"},
		},
		{
			name: "fenced code with litdoc",
			block: internal.MakeBlockFromRaw(
				internal.BlockKindFencedCode,
				[]byte("```bash | litdoc\necho hello\n```\n"),
			),
			want: internal.InfoString{Lang: "bash", IsLitdoc: true},
		},
		{
			name: "html comment with litdoc",
			block: internal.MakeBlockFromRaw(
				internal.BlockKindHTMLComment,
				[]byte("<!-- bash | litdoc\necho hello\n-->\n"),
			),
			want: internal.InfoString{Lang: "bash", IsLitdoc: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := internal.ParseInfoString(tt.block)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMakeBashCellFromRaw(t *testing.T) {
	// given
	code := "```bash\necho hello\n```\n"

	// when
	gotCell := internal.MakeBashCellFromRaw(code, internal.Output{})

	// then
	got, err := gotCell.Render()
	require.NoError(t, err)
	assert.Equal(t, code, got)
}

func TestBashCellExecute(t *testing.T) {
	// given
	fencedCode := "```bash\necho hello\n```\n"
	cell := internal.MakeBashCellFromRaw(fencedCode, internal.Output{})

	// when
	gotCell, err := cell.Execute()

	// then
	require.NoError(t, err)
	rendered, err := gotCell.Render()
	require.NoError(t, err)
	want := fencedCode + "\n<!-- BEGIN litdoc OUTPUT -->\noutput\n<!-- END litdoc OUTPUT -->\n"
	assert.Equal(t, want, rendered)
}

func TestBashCellRender(t *testing.T) {
	t.Run("without output", func(t *testing.T) {
		// given
		code := "```bash\necho hello\n```\n"
		cell := internal.MakeBashCellFromRaw(code, internal.Output{})

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, code, gotContent)
	})

	t.Run("with output", func(t *testing.T) {
		// given
		fencedCode := "```bash\necho hello\n```\n"
		cell := internal.MakeBashCellFromRaw(fencedCode, internal.Output{})
		executed, err := cell.Execute()
		require.NoError(t, err)

		// when
		gotContent, err := executed.Render()

		// then
		require.NoError(t, err)
		want := fencedCode + "\n<!-- BEGIN litdoc OUTPUT -->\noutput\n<!-- END litdoc OUTPUT -->\n"
		assert.Equal(t, want, gotContent)
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
	textBlock := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindText, []byte(content))
	}
	bashLitdocBlock := func(content string) internal.Block {
		return internal.MakeBlockFromRaw(internal.BlockKindFencedCode, []byte(content))
	}

	t.Run("single text block becomes StaticCell", func(t *testing.T) {
		// given
		blocks := []internal.Block{textBlock("hello")}

		// when
		cells, err := internal.Classify(blocks)

		// then
		require.NoError(t, err)
		require.Len(t, cells, 1)
		_, ok := cells[0].(internal.StaticCell)
		require.True(t, ok, "expected StaticCell, got %T", cells[0])
		got, err := cells[0].Render()
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
	})

	t.Run("litdoc bash block becomes BashCell", func(t *testing.T) {
		// given
		code := "```bash | litdoc\necho hello\n```\n"
		blocks := []internal.Block{bashLitdocBlock(code)}

		// when
		cells, err := internal.Classify(blocks)

		// then
		require.NoError(t, err)
		require.Len(t, cells, 1)
		_, ok := cells[0].(internal.BashCell)
		require.True(t, ok, "expected BashCell, got %T", cells[0])
		rendered, err := cells[0].Render()
		require.NoError(t, err)
		assert.Equal(t, code, rendered)
	})

	t.Run("mixed block types are each classified independently", func(t *testing.T) {
		// given
		code := "```bash | litdoc\necho hello\n```\n"
		blocks := []internal.Block{
			textBlock("before"),
			bashLitdocBlock(code),
			textBlock("after"),
		}

		// when
		cells, err := internal.Classify(blocks)

		// then
		require.NoError(t, err)
		require.Len(t, cells, 3)
		rendered0, err := cells[0].Render()
		require.NoError(t, err)
		assert.Equal(t, "before", rendered0)
		_, ok := cells[1].(internal.BashCell)
		assert.True(t, ok, "expected BashCell, got %T", cells[1])
		rendered2, err := cells[2].Render()
		require.NoError(t, err)
		assert.Equal(t, "after", rendered2)
	})

	t.Run("non-litdoc fenced code block becomes StaticCell", func(t *testing.T) {
		// given
		code := "```bash\necho hello\n```\n"
		blocks := []internal.Block{
			internal.MakeBlockFromRaw(internal.BlockKindFencedCode, []byte(code)),
		}

		// when
		cells, err := internal.Classify(blocks)

		// then
		require.NoError(t, err)
		require.Len(t, cells, 1)
		rendered, err := cells[0].Render()
		require.NoError(t, err)
		assert.Equal(t, code, rendered)
	})

	t.Run("output block following bash cell is loaded into BashCell output", func(t *testing.T) {
		// given - mirrors what tree-sitter produces: three separate blocks for the output section
		code := "```bash | litdoc\necho hello\n```\n"
		blocks := []internal.Block{
			internal.MakeBlockFromRaw(internal.BlockKindFencedCode, []byte(code)),
			internal.MakeBlockFromRaw(internal.BlockKindText, []byte("\n")),
			internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, []byte("<!-- BEGIN litdoc OUTPUT -->\n")),
			internal.MakeBlockFromRaw(internal.BlockKindText, []byte("old output\n")),
			internal.MakeBlockFromRaw(internal.BlockKindHTMLComment, []byte("<!-- END litdoc OUTPUT -->\n")),
		}

		// when
		cells, err := internal.Classify(blocks)

		// then
		require.NoError(t, err)
		require.Len(t, cells, 1)
		_, ok := cells[0].(internal.BashCell)
		require.True(t, ok, "expected BashCell, got %T", cells[0])
		rendered, err := cells[0].Render()
		require.NoError(t, err)
		wantOutput := "<!-- BEGIN litdoc OUTPUT -->\nold output\n<!-- END litdoc OUTPUT -->\n"
		assert.Equal(t, code+"\n"+wantOutput, rendered)
	})

	t.Run("litdoc block with unsupported language", func(t *testing.T) {
		// given
		blocks := []internal.Block{
			bashLitdocBlock("```go | litdoc\nfmt.Println()\n```\n"),
		}

		// when
		_, err := internal.Classify(blocks)

		// then
		require.ErrorContains(t, err, "unsupported language")
	})
}
