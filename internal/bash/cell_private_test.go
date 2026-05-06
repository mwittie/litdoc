package bash

import (
	"testing"

	"litdoc/internal"

	"github.com/stretchr/testify/assert"
)

func TestMakeCellFromRaw(t *testing.T) {
	// given
	content := "```bash\necho hello\n```\n"
	indent := "  "
	output := internal.MakeOutput("hello", "")
	runner := Runner{}
	block := internal.MakeBlockFromRaw(internal.BlockKindFencedCode, content, indent, false)

	// when
	c := makeCellFromRaw(content, indent, output, runner)

	// then
	assert.Equal(t, block, c.block)
	assert.Equal(t, output, c.output)
	assert.Equal(t, runner, c.runner)
}

func TestMakeParserFromRaw(t *testing.T) {
	// given
	runner := Runner{}

	// when
	p := makeParserFromRaw(runner, internal.OutputParserFunc(internal.OutputFromBlocks))

	// then
	assert.Equal(t, runner, p.runner)
	assert.NotNil(t, p.outputParser)
}
