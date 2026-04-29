package internal

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

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
	indent  []byte
	content []byte
}

func MakeBlockFromRaw(kind BlockKind, raw []byte) Block {
	return Block{kind: kind, content: raw}
}

func makeBlock(kind BlockKind, content, indent []byte) Block {
	return Block{kind: kind, indent: indent, content: content}
}

func (b Block) Kind() BlockKind { return b.kind }

func (b Block) Indent() []byte { return b.indent }

func (b Block) Content() []byte { return b.content }

func (b Block) String() string {
	return fmt.Sprintf("{%s %q}", b.kind, b.content)
}

var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

func MakeBlocksFromMarkdown(content []byte) ([]Block, error) {
	tree, err := markdown.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown: %w", err)
	}

	root := tree.BlockTree().RootNode()
	var blocks []Block
	pos := uint32(0)

	collectBlockNodes(root, content, nil, &pos, &blocks)

	if pos < uint32(len(content)) {
		blocks = append(blocks, makeBlock(BlockKindText, content[pos:], nil))
	}

	return splitInlineHTMLComments(blocks), nil
}

func splitInlineHTMLComments(blocks []Block) []Block {
	result := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Kind() != BlockKindText {
			result = append(result, b)
			continue
		}
		content := b.Content()
		locs := htmlCommentRe.FindAllIndex(content, -1)
		if len(locs) == 0 {
			result = append(result, b)
			continue
		}
		pos := 0
		for _, loc := range locs {
			if loc[0] > pos {
				result = append(result, makeBlock(BlockKindText, content[pos:loc[0]], b.indent))
			}
			result = append(result, makeBlock(BlockKindHTMLComment, content[loc[0]:loc[1]], b.indent))
			pos = loc[1]
		}
		if pos < len(content) {
			result = append(result, makeBlock(BlockKindText, content[pos:], b.indent))
		}
	}
	return result
}

func collectBlockNodes(
	node *sitter.Node,
	content []byte,
	indent []byte,
	pos *uint32,
	blocks *[]Block,
) {
	switch node.Type() {
	case "document", "section":
		for i := 0; i < int(node.ChildCount()); i++ {
			collectBlockNodes(node.Child(i), content, indent, pos, blocks)
		}
	case "block_quote":
		childIndent := append(append([]byte(nil), indent...), "> "...)
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "block_quote_marker" {
				if child.StartByte() > *pos {
					gap := stripIndent(content[*pos:child.StartByte()], childIndent)
					if len(gap) > 0 {
						*blocks = append(*blocks, makeBlock(BlockKindText, gap, childIndent))
					}
				}
				*pos = child.EndByte()
				continue
			}
			collectBlockNodes(child, content, childIndent, pos, blocks)
		}
	default:
		start := node.StartByte()
		end := node.EndByte()

		if start > *pos {
			gap := stripIndent(content[*pos:start], indent)
			if len(gap) > 0 {
				*blocks = append(*blocks, makeBlock(BlockKindText, gap, indent))
			}
		}

		raw := stripIndent(content[start:end], indent)
		if len(raw) > 0 {
			*blocks = append(*blocks, makeBlock(blockKind(node, content), raw, indent))
		}
		*pos = end
	}
}

func stripIndent(content, indent []byte) []byte {
	if len(indent) == 0 {
		return content
	}
	lines := bytes.Split(content, []byte("\n"))
	parentIndent := quoteParentIndent(indent)
	for i, line := range lines {
		switch {
		case bytes.HasPrefix(line, indent):
			lines[i] = bytes.TrimPrefix(line, indent)
		case len(parentIndent) > 0 && bytes.Equal(line, parentIndent):
			lines[i] = nil
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

func quoteParentIndent(indent []byte) []byte {
	if len(indent) < len("> ") || !bytes.HasSuffix(indent, []byte("> ")) {
		return nil
	}
	return indent[:len(indent)-len("> ")]
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
