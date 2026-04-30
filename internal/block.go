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
	BlockKindText        BlockKind = "text"
	BlockKindFencedCode  BlockKind = "fenced_code"
	BlockKindHTMLComment BlockKind = "html_comment"
)

type Block struct {
	kind BlockKind
	// indent characters potentially nested blockquotes and lists
	indent  string
	content string
	// continuation when preceding block is on the smame line
	continuation bool
}

func MakeBlockFromRaw(
	kind BlockKind,
	content string,
	indent string,
	continuation bool,
) Block {
	return Block{
		kind:         kind,
		indent:       indent,
		content:      content,
		continuation: continuation,
	}
}

func (b Block) Kind() BlockKind { return b.kind }

func (b Block) Indent() string { return b.indent }

func (b Block) Content() string { return b.content }

func (b Block) Continuation() bool { return b.continuation }

func (b Block) String() string {
	return fmt.Sprintf("{%s %q}", b.kind, b.content)
}

func MakeBlocksFromMarkdown(content []byte) ([]Block, error) {
	if !utf8.Valid(content) {
		return nil, fmt.Errorf("markdown content is not valid UTF-8")
	}

	tree, err := markdown.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("parsing markdown: %w", err)
	}

	root := tree.BlockTree().RootNode()
	collector := blockCollector{content: content}
	collector.collect(root, nil, nil)
	if collector.err != nil {
		return nil, collector.err
	}
	collector.appendTrailingText()

	// inline comments come in a text block. Split them out in a linear pass.
	return splitInlineHTMLComments(collector.blocks), nil
}

type blockCollector struct {
	content []byte
	pos     uint32
	blocks  []Block
	err     error
}

func (c *blockCollector) collect(node *sitter.Node, indent, stripPrefix []byte) {
	if c.err != nil {
		return
	}

	switch node.Type() {
	case "document", "section", "list":
		c.collectChildren(node, indent, stripPrefix)
	case "block_quote":
		c.collectBlockQuote(node, indent, stripPrefix)
	case "list_item":
		c.collectListItem(node, indent, stripPrefix)
	default:
		c.collectLeaf(node, indent, stripPrefix)
	}
}

func (c *blockCollector) collectChildren(node *sitter.Node, indent, stripPrefix []byte) {
	for i := 0; i < int(node.ChildCount()); i++ {
		c.collect(node.Child(i), indent, stripPrefix)
	}
}

func (c *blockCollector) collectBlockQuote(
	node *sitter.Node,
	indent []byte,
	stripPrefix []byte,
) {
	childIndent := concatenate(indent, []byte("> "))
	childStripPrefix := concatenate(stripPrefix, []byte("> "))
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "block_quote_marker" {
			c.appendTextGap(child.StartByte(), childStripPrefix, childIndent)
			c.pos = child.EndByte()
			continue
		}
		c.collect(child, childIndent, childStripPrefix)
	}
}

func (c *blockCollector) collectListItem(node *sitter.Node, indent, stripPrefix []byte) {
	childIndent, childStripPrefix := listItemPrefixes(node, c.content)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if isListMarker(child) {
			c.appendTextGap(child.StartByte(), stripPrefix, indent)
			c.pos = child.EndByte()
			continue
		}
		c.collect(child, childIndent, childStripPrefix)
	}
}

func (c *blockCollector) collectLeaf(node *sitter.Node, indent, stripPrefix []byte) {
	start := node.StartByte()
	end := node.EndByte()
	blockIndent, blockStripPrefix := blockPrefixes(node, c.content, indent, stripPrefix)

	c.appendTextGap(start, stripPrefix, indent)

	raw := stripIndent(c.content[start:end], blockStripPrefix)
	kind := blockKind(node, c.content)
	if kind == BlockKindFencedCode && !isClosedFencedCodeBlock(raw) {
		c.err = fmt.Errorf("unclosed fenced code block at byte %d", start)
		return
	}
	if kind == BlockKindHTMLComment && !isClosedHTMLComment(raw) {
		c.err = fmt.Errorf("unclosed HTML comment at byte %d", start)
		return
	}

	c.appendBlock(kind, raw, blockIndent)
	c.pos = end
}

func (c *blockCollector) appendTrailingText() {
	if c.pos < uint32(len(c.content)) {
		c.appendBlock(BlockKindText, c.content[c.pos:], nil)
	}
}

func (c *blockCollector) appendTextGap(end uint32, stripPrefix, indent []byte) {
	if end <= c.pos {
		return
	}
	gap := stripIndent(c.content[c.pos:end], stripPrefix)
	c.appendBlock(BlockKindText, gap, indent)
}

func (c *blockCollector) appendBlock(kind BlockKind, raw, indent []byte) {
	if len(raw) == 0 {
		return
	}
	if kind == BlockKindText && isBlockQuoteOnlyIndent(indent) && !bytes.Contains(raw, []byte("<!--")) {
		c.appendBlockQuoteText(raw, indent)
		return
	}
	c.blocks = append(c.blocks, MakeBlockFromRaw(kind, string(raw), string(indent), false))
}

func (c *blockCollector) appendBlockQuoteText(raw, indent []byte) {
	for len(raw) > 0 {
		lineEnd := bytes.IndexByte(raw, '\n')
		if lineEnd < 0 {
			c.blocks = append(c.blocks, MakeBlockFromRaw(
				BlockKindText,
				string(raw),
				string(indent),
				false,
			))
			return
		}
		lineEnd++
		c.blocks = append(c.blocks, MakeBlockFromRaw(
			BlockKindText,
			string(raw[:lineEnd]),
			string(indent),
			false,
		))
		raw = raw[lineEnd:]
	}
}

func isBlockQuoteOnlyIndent(indent []byte) bool {
	if !bytes.Contains(indent, []byte("> ")) {
		return false
	}
	for len(indent) > 0 {
		for len(indent) > 0 && indent[0] == ' ' {
			indent = indent[1:]
		}
		if !bytes.HasPrefix(indent, []byte("> ")) {
			return false
		}
		indent = indent[len("> "):]
	}
	return true
}

func blockPrefixes(
	node *sitter.Node,
	content []byte,
	indent []byte,
	stripPrefix []byte,
) ([]byte, []byte) {
	start := node.StartByte()
	end := node.EndByte()
	blockIndent := indent
	blockStripPrefix := stripPrefix

	// Tree-sitter nodes start after some markdown prefixes but include others.
	// When a node begins after a prefix that differs from the current container,
	// strip that prefix from stored content and keep it as the render indent.
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
		return blockIndent, blockStripPrefix
	}

	if node.Type() != "fenced_code_block" || len(indent) > 0 {
		return blockIndent, blockStripPrefix
	}

	// CommonMark allows top-level fenced code blocks to be indented up to three
	// spaces. Treat that as the block prefix rather than code content.
	rawLeading := leadingLineSpaces(content[start:end])
	if n := len(rawLeading); n > 0 && n <= 3 {
		blockIndent = rawLeading
		blockStripPrefix = rawLeading
	}
	return blockIndent, blockStripPrefix
}

func isClosedFencedCodeBlock(content []byte) bool {
	openLineEnd := bytes.IndexByte(content, '\n')
	if openLineEnd < 0 {
		return false
	}

	fenceChar, fenceLen, ok := openingFence(content[:openLineEnd])
	if !ok {
		return false
	}

	rest := content[openLineEnd+1:]
	if len(rest) == 0 {
		return false
	}
	lastLine := lastContentLine(rest)
	if len(lastLine) == 0 {
		return false
	}
	return isClosingFence(lastLine, fenceChar, fenceLen)
}

var openingFenceRe = regexp.MustCompile(`^[ ]{0,3}(` + "`" + `{3,}|~{3,})`)

func openingFence(line []byte) (byte, int, bool) {
	line = bytes.TrimSuffix(line, []byte("\r"))
	match := openingFenceRe.FindSubmatch(line)
	if match == nil {
		return 0, 0, false
	}
	fence := match[1]
	fenceChar := fence[0]
	return fenceChar, len(fence), true
}

func lastContentLine(content []byte) []byte {
	content = bytes.TrimSuffix(content, []byte("\n"))
	if i := bytes.LastIndexByte(content, '\n'); i >= 0 {
		return content[i+1:]
	}
	return content
}

func isClosingFence(line []byte, fenceChar byte, fenceLen int) bool {
	fence, ok := trimFenceLineIndent(line)
	if !ok || len(fence) < fenceLen {
		return false
	}

	closingLen := countLeadingByte(fence, fenceChar)
	if closingLen < fenceLen {
		return false
	}

	return len(bytes.Trim(fence[closingLen:], " \t")) == 0
}

func trimFenceLineIndent(line []byte) ([]byte, bool) {
	line = bytes.TrimSuffix(line, []byte("\r"))
	n := countLeadingByte(line, ' ')
	if n > 3 {
		return nil, false
	}
	return line[n:], true
}

func countLeadingByte(content []byte, b byte) int {
	n := 0
	for n < len(content) && content[n] == b {
		n++
	}
	return n
}

func isClosedHTMLComment(content []byte) bool {
	return bytes.Contains(content, []byte("-->"))
}

func concatenate(left, right []byte) []byte {
	result := make([]byte, 0, len(left)+len(right))
	result = append(result, left...)
	result = append(result, right...)
	return result
}

// listItemPrefixes returns both views of a list marker: the marker itself is
// used when rendering the first line, while equivalent spaces are stripped from
// continuation lines in the source.
func listItemPrefixes(node *sitter.Node, content []byte) ([]byte, []byte) {
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
	return strings.HasPrefix(node.Type(), "list_marker_")
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

// stripIndent removes a container prefix from each line before storing content
// in a Block. Blockquote-only blank lines are normalized to empty lines so the
// rendered output does not keep stray ">" markers as content.
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

var htmlCommentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

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
				result = append(
					result,
					MakeBlockFromRaw(
						BlockKindText,
						content[pos:loc[0]],
						b.indent,
						b.continuation || pos > 0,
					),
				)
			}
			wholeBlock := isWholeTextBlock(content, loc)
			commentContinuation := b.continuation
			if !wholeBlock {
				commentContinuation = b.continuation || loc[0] > 0
			}
			result = append(result, MakeBlockFromRaw(
				BlockKindHTMLComment,
				content[loc[0]:loc[1]],
				b.indent,
				commentContinuation,
			))
			pos = loc[1]
		}
		if pos < len(content) {
			result = append(
				result,
				MakeBlockFromRaw(BlockKindText, content[pos:], b.indent, true),
			)
		}
	}
	return result
}

func isWholeTextBlock(content string, loc []int) bool {
	return len(strings.TrimSpace(content[:loc[0]])) == 0 &&
		len(strings.TrimSpace(content[loc[1]:])) == 0
}
