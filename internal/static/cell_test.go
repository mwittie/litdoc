package static_test

import (
	"testing"

	"litdoc/internal/static"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticCell(t *testing.T) {
	t.Run("renders raw content", func(t *testing.T) {
		// given
		cell := static.MakeStaticCellFromRaw("hello")

		// when
		gotContent, err := cell.Render()

		// then
		require.NoError(t, err)
		assert.Equal(t, "hello", gotContent)
	})

	t.Run("executes to itself", func(t *testing.T) {
		// given
		cell := static.MakeStaticCellFromRaw("hello")

		// when
		gotCell, err := cell.Execute()

		// then
		require.NoError(t, err)
		assert.Equal(t, cell, gotCell)
	})
}

