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
	indent  string
}

func MakeBlockFromRaw(kind BlockKind, raw []byte, indent string) Block {
	return Block{kind: kind, content: raw, indent: indent}
}

func (b Block) Kind() BlockKind { return b.kind }
func (b Block) Content() []byte { return b.content }
func (b Block) Indent() string  { return b.indent }

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
		blocks = append(blocks, MakeBlockFromRaw(BlockKindText, content[pos:], ""))
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
	case "document", "section", "list", "list_item":
		for i := 0; i < int(node.ChildCount()); i++ {
			collectBlockNodes(node.Child(i), content, pos, blocks)
		}
	default:
		start := node.StartByte()
		end := node.EndByte()

		if start > *pos {
			*blocks = append(*blocks, MakeBlockFromRaw(BlockKindText, content[*pos:start], ""))
		}

		kind := blockKind(node, content)
		indent := ""
		if kind == BlockKindFencedCode || kind == BlockKindHTMLComment {
			indent = lineIndentBefore(start, content)
		}
		*blocks = append(*blocks, MakeBlockFromRaw(kind, content[start:end], indent))
		*pos = end
	}
}

// lineIndentBefore returns the leading whitespace on the line containing start.
func lineIndentBefore(start uint32, content []byte) string {
	lineStart := start
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	i := lineStart
	for i < start && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	return string(content[lineStart:i])
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
