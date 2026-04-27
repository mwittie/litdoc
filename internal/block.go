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
		blocks = append(blocks, MakeBlockFromRaw(BlockKindText, content[pos:]))
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
		kind := blockKind(node, content)
		blockStart := start
		blockEnd := lineTrailingWhitespaceStart(end, content)
		if kind == BlockKindFencedCode || kind == BlockKindHTMLComment {
			blockStart = lineWhitespaceStart(start, content)
			if blockStart < *pos && len(*blocks) > 0 {
				last := len(*blocks) - 1
				if (*blocks)[last].kind == BlockKindText &&
					bytes.Equal((*blocks)[last].content, content[blockStart:*pos]) {
					*blocks = (*blocks)[:last]
					*pos = blockStart
				}
			}
			if blockStart < *pos {
				blockStart = start
			}
		}

		if blockStart > *pos {
			if kind == BlockKindText && !isListMarker(content[blockStart:blockEnd]) {
				*blocks = append(*blocks, MakeBlockFromRaw(BlockKindText, content[*pos:blockEnd]))
				*pos = blockEnd
				return
			}
			*blocks = append(*blocks, MakeBlockFromRaw(BlockKindText, content[*pos:blockStart]))
		}

		*blocks = append(*blocks, MakeBlockFromRaw(kind, content[blockStart:blockEnd]))
		*pos = blockEnd
	}
}

func lineWhitespaceStart(start uint32, content []byte) uint32 {
	lineStart := start
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	i := lineStart
	for i < start && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	if i == start {
		return lineStart
	}
	return start
}

func lineTrailingWhitespaceStart(end uint32, content []byte) uint32 {
	i := end
	for i > 0 && (content[i-1] == ' ' || content[i-1] == '\t') {
		i--
	}
	if i > 0 && content[i-1] == '\n' {
		return i
	}
	return end
}

func isListMarker(content []byte) bool {
	if bytes.Equal(content, []byte("- ")) ||
		bytes.Equal(content, []byte("+ ")) ||
		bytes.Equal(content, []byte("* ")) {
		return true
	}
	if len(content) < 3 || content[len(content)-1] != ' ' {
		return false
	}
	for i, b := range content[:len(content)-2] {
		if i == 0 && (b < '0' || b > '9') {
			return false
		}
		if i > 0 && (b < '0' || b > '9') {
			return false
		}
	}
	marker := content[len(content)-2]
	return marker == '.' || marker == ')'
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
