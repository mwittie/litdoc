package internal

import (
	"bytes"
	"context"
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/markdown"
)

type BlockKind string

const (
	BlockKindText        BlockKind = "Text"
	BlockKindFencedCode  BlockKind = "fencedCode"
	BlockKindHTMLComment BlockKind = "HTMLComment"
)

type Block struct {
	kind    BlockKind
	content []byte
}

func MakeBlockFromRaw(kind BlockKind, raw []byte) Block {
	return Block{kind: kind, content: raw}
}

func (b Block) Kind() BlockKind { return b.kind }

func (b Block) Content() []byte { return b.content }

func MakeBlocksFromMarkdown(content []byte) ([]Block, error) {
	tree, err := markdown.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown: %w", err)
	}

	root := tree.BlockTree().RootNode()
	var blocks []Block
	pos := uint32(0)

	collectBlockNodes(root, content, &pos, &blocks)

	if pos < uint32(len(content)) {
		blocks = append(blocks, Block{kind: BlockKindText, content: content[pos:]})
	}

	return blocks, nil
}

func collectBlockNodes(
	node *sitter.Node,
	content []byte,
	pos *uint32,
	blocks *[]Block,
) {
	switch node.Type() {
	case "document", "section":
		for i := 0; i < int(node.ChildCount()); i++ {
			collectBlockNodes(node.Child(i), content, pos, blocks)
		}
	default:
		start := node.StartByte()
		end := node.EndByte()

		if start > *pos {
			*blocks = append(
				*blocks,
				MakeBlockFromRaw(BlockKindText, content[*pos:start]),
			)
		}

		*blocks = append(
			*blocks,
			MakeBlockFromRaw(blockKind(node, content), content[start:end]),
		)
		*pos = end
	}
}

func blockKind(node *sitter.Node, content []byte) BlockKind {
	switch node.Type() {
	case "fenced_code_block":
		return BlockKindFencedCode
	case "html_block":
		if bytes.HasPrefix(content[node.StartByte():node.EndByte()], []byte("<!--")) {
			return BlockKindHTMLComment
		}
		return BlockKindText
	default:
		return BlockKindText
	}
}
