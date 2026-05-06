package static

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
)

func TestMakeCellFromRaw(t *testing.T) {
	// given
	content := "hello"
	block := internal.MakeBlockFromRaw(internal.BlockKindText, content, "", false)

	// when
	c := makeCellFromRaw(content)

	// then
	assert.Equal(t, block, c.block)
}
