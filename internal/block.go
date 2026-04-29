package internal

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

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
	indent  string
	content string
}

func MakeBlock(kind BlockKind, content, indent string) Block {
	return Block{kind: kind, indent: indent, content: content}
}

func (b Block) Kind() BlockKind { return b.kind }

func (b Block) Indent() string { return b.indent }

func (b Block) Content() string { return b.content }

func (b Block) String() string {
	return fmt.Sprintf("{%s %q}", b.kind, b.content)
}

var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

func MakeBlocksFromMarkdown(content []byte) ([]Block, error) {
	if !utf8.Valid(content) {
		return nil, fmt.Errorf("markdown content is not valid UTF-8")
	}

	tree, err := markdown.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown: %w", err)
	}

	root := tree.BlockTree().RootNode()
	var blocks []Block
	pos := uint32(0)

	collectBlockNodes(root, content, nil, nil, &pos, &blocks)

	if pos < uint32(len(content)) {
		blocks = append(blocks, MakeBlock(BlockKindText, string(content[pos:]), ""))
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
		locs := htmlCommentRe.FindAllStringIndex(content, -1)
		if len(locs) == 0 {
			result = append(result, b)
			continue
		}
		pos := 0
		for _, loc := range locs {
			if loc[0] > pos {
				result = append(result, MakeBlock(BlockKindText, content[pos:loc[0]], b.indent))
			}
			commentIndent := ""
			if isWholeTextBlock(content, loc) {
				commentIndent = b.indent
			}
			result = append(result, MakeBlock(BlockKindHTMLComment, content[loc[0]:loc[1]], commentIndent))
			pos = loc[1]
		}
		if pos < len(content) {
			result = append(result, MakeBlock(BlockKindText, content[pos:], b.indent))
		}
	}
	return result
}

func isWholeTextBlock(content string, loc []int) bool {
	return len(strings.TrimSpace(content[:loc[0]])) == 0 &&
		len(strings.TrimSpace(content[loc[1]:])) == 0
}

func collectBlockNodes(
	node *sitter.Node,
	content []byte,
	indent []byte,
	stripPrefix []byte,
	pos *uint32,
	blocks *[]Block,
) {
	switch node.Type() {
	case "document", "section":
		for i := 0; i < int(node.ChildCount()); i++ {
			collectBlockNodes(node.Child(i), content, indent, stripPrefix, pos, blocks)
		}
	case "block_quote":
		childIndent := append(append([]byte(nil), indent...), "> "...)
		childStripPrefix := append(append([]byte(nil), stripPrefix...), "> "...)
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "block_quote_marker" {
				if child.StartByte() > *pos {
					gap := stripIndent(content[*pos:child.StartByte()], childStripPrefix)
					if len(gap) > 0 {
						*blocks = append(*blocks, MakeBlock(BlockKindText, string(gap), string(childIndent)))
					}
				}
				*pos = child.EndByte()
				continue
			}
			collectBlockNodes(child, content, childIndent, childStripPrefix, pos, blocks)
		}
	case "list":
		for i := 0; i < int(node.ChildCount()); i++ {
			collectBlockNodes(node.Child(i), content, indent, stripPrefix, pos, blocks)
		}
	case "list_item":
		childIndent, childStripPrefix := listItemIndent(node, content)
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if isListMarker(child) {
				if child.StartByte() > *pos {
					gap := stripIndent(content[*pos:child.StartByte()], stripPrefix)
					if len(gap) > 0 {
						*blocks = append(*blocks, MakeBlock(BlockKindText, string(gap), string(indent)))
					}
				}
				*pos = child.EndByte()
				continue
			}
			collectBlockNodes(child, content, childIndent, childStripPrefix, pos, blocks)
		}
	default:
		start := node.StartByte()
		end := node.EndByte()
		blockIndent := indent
		blockStripPrefix := stripPrefix
		linePrefix := linePrefixBefore(content, start)
		if len(linePrefix) > 0 && !bytes.Equal(linePrefix, indent) {
			blockIndent = linePrefix
			blockStripPrefix = linePrefix
			if isSpaceIndent(linePrefix) {
				rawLeading := leadingLineSpaces(content[start:end])
				if len(rawLeading) > 0 {
					blockIndent = append(append([]byte(nil), linePrefix...), rawLeading...)
					blockStripPrefix = blockIndent
				}
			}
		} else if node.Type() == "fenced_code_block" && len(indent) == 0 {
			// CommonMark §4.5: a fence may be indented up to 3 spaces; that indent is
			// part of the block's prefix, not the code content.
			rawLeading := leadingLineSpaces(content[start:end])
			if n := len(rawLeading); n > 0 && n <= 3 {
				blockIndent = rawLeading
				blockStripPrefix = rawLeading
			}
		}

		if start > *pos {
			gap := stripIndent(content[*pos:start], stripPrefix)
			if len(gap) > 0 {
				*blocks = append(*blocks, MakeBlock(BlockKindText, string(gap), string(indent)))
			}
		}

		raw := stripIndent(content[start:end], blockStripPrefix)
		if len(raw) > 0 {
			*blocks = append(*blocks, MakeBlock(blockKind(node, content), string(raw), string(blockIndent)))
		}
		*pos = end
	}
}

func listItemIndent(node *sitter.Node, content []byte) ([]byte, []byte) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if !isListMarker(child) {
			continue
		}
		linePrefix := linePrefixBefore(content, child.StartByte())
		marker := content[child.StartByte():child.EndByte()]
		indent := append(append([]byte(nil), linePrefix...), marker...)
		stripPrefix := append(append([]byte(nil), linePrefix...), bytes.Repeat([]byte(" "), len(marker))...)
		return indent, stripPrefix
	}
	return nil, nil
}

func isListMarker(node *sitter.Node) bool {
	return bytes.HasPrefix([]byte(node.Type()), []byte("list_marker_"))
}

func linePrefixBefore(content []byte, pos uint32) []byte {
	lineStart := bytes.LastIndexByte(content[:pos], '\n') + 1
	return content[lineStart:pos]
}

func leadingLineSpaces(content []byte) []byte {
	end := bytes.IndexByte(content, '\n')
	if end < 0 {
		end = len(content)
	}
	line := content[:end]
	return line[:len(line)-len(bytes.TrimLeft(line, " "))]
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
		case isSpaceIndent(indent):
			lines[i] = trimSpaceIndent(line, len(indent))
		case bytes.Contains(indent, []byte("> ")) && isQuoteOnlyLine(line):
			lines[i] = nil
		case len(parentIndent) > 0 && bytes.Equal(line, parentIndent):
			lines[i] = nil
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

func isSpaceIndent(indent []byte) bool {
	return len(bytes.Trim(indent, " ")) == 0
}

func trimSpaceIndent(line []byte, width int) []byte {
	i := 0
	for i < len(line) && i < width && line[i] == ' ' {
		i++
	}
	return line[i:]
}

func isQuoteOnlyLine(line []byte) bool {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}
	for _, b := range trimmed {
		if b != '>' && b != ' ' {
			return false
		}
	}
	return true
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
