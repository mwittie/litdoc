package static_test

import (
	"testing"

	"litdoc/internal"
	"litdoc/internal/static"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeCellFromRaw(t *testing.T) {
	// given
	cell := static.MakeCellFromRaw("hello")

	// when
	got, err := cell.Render()

	// then
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestParse(t *testing.T) {
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

func TestRender(t *testing.T) {
	// given
	cell := static.MakeCellFromRaw("hello")

	// when
	got, err := cell.Render()

	// then
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestExecute(t *testing.T) {
	// given
	cell := static.MakeCellFromRaw("hello")

	// when
	got, err := cell.Execute()

	// then
	require.NoError(t, err)
	assert.Equal(t, cell, got)
}
