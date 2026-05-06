package static_test

import (
	"testing"

	"litdoc/internal"
	"litdoc/internal/static"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestParser_Parse(t *testing.T) {
	// given
	block := internal.MakeBlockFromRaw(internal.BlockKindText, "hello", "", false)

	// when
	cell, consumed, err := static.NewParser().Parse(block, nil)

	// then
	require.NoError(t, err)
	assert.Equal(t, 0, consumed)
	rendered, err := cell.Render()
	require.NoError(t, err)
	assert.Equal(t, block.Render(), rendered)
}

func TestCell_Render(t *testing.T) {
	// given
	cell := static.MakeCellFromRaw("hello")

	// when
	got, err := cell.Render()

	// then
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestCell_Execute(t *testing.T) {
	// given
	cell := static.MakeCellFromRaw("hello")

	// when
	got, err := cell.Execute()

	// then
	require.NoError(t, err)
	assert.Equal(t, cell, got)
}
