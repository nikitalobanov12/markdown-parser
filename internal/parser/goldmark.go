package parser

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	"markdown-parser/internal/models"
)

// MarkdownParser wraps Goldmark with additional functionality
type MarkdownParser struct {
	goldmark goldmark.Markdown
}

// NewMarkdownParser creates a new parser with GitHub Flavored Markdown extensions
func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // GitHub Flavored Markdown
			extension.Footnote,      // Footnote support
			extension.DefinitionList, // Definition list support
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),     // Convert line breaks to <br>
			html.WithXHTML(),         // Use XHTML-style output
			html.WithUnsafe(),        // Allow raw HTML
		),
	)

	return &MarkdownParser{
		goldmark: md,
	}
}

// Parse converts markdown to HTML and extracts block information
func (p *MarkdownParser) Parse(content string) (*models.ParseResponse, error) {
	if content == "" {
		return &models.ParseResponse{
			HTML:    "",
			Blocks:  make(map[string]*models.Block),
			Success: true,
		}, nil
	}

	// Parse to HTML
	var htmlBuf bytes.Buffer
	source := []byte(content)
	
	doc := p.goldmark.Parser().Parse(text.NewReader(source))
	if err := p.goldmark.Renderer().Render(&htmlBuf, source, doc); err != nil {
		return nil, fmt.Errorf("failed to render HTML: %w", err)
	}

	// Extract blocks from AST
	blocks := p.extractBlocks(doc, source)

	return &models.ParseResponse{
		HTML:    htmlBuf.String(),
		Blocks:  blocks,
		Success: true,
	}, nil
}

// ParseIncremental performs incremental parsing for real-time updates
func (p *MarkdownParser) ParseIncremental(content string, blockID string) (*models.ParseResponse, error) {
	// For now, we'll parse the entire content
	// In a production system, you'd implement proper incremental parsing
	return p.Parse(content)
}

// extractBlocks walks the AST and extracts block information
func (p *MarkdownParser) extractBlocks(doc ast.Node, source []byte) map[string]*models.Block {
	blocks := make(map[string]*models.Block)
	
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		block := p.nodeToBlock(n, source)
		if block != nil {
			blocks[block.ID] = block
		}

		return ast.WalkContinue, nil
	})

	return blocks
}

// nodeToBlock converts an AST node to a Block
func (p *MarkdownParser) nodeToBlock(node ast.Node, source []byte) *models.Block {
	// Only process block-level elements
	kind := node.Kind()
	if kind == ast.KindText || kind == ast.KindString {
		return nil
	}

	var startPos, endPos int
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		segment := hasSegment.Segment()
		startPos = segment.Start
		endPos = segment.Stop
	}

	block := &models.Block{
		ID:       p.generateBlockID(node, source),
		Position: models.Position{
			Start: startPos,
			End:   endPos,
		},
	}

	// Extract content from source
	if startPos < len(source) && endPos <= len(source) && endPos > startPos {
		block.Content = string(source[startPos:endPos])
	}

	// Determine block type and extract relevant information
	switch n := node.(type) {
	case *ast.Heading:
		switch n.Level {
		case 1:
			block.Type = "h1"
		case 2:
			block.Type = "h2"
		case 3:
			block.Type = "h3"
		case 4:
			block.Type = "h4"
		case 5:
			block.Type = "h5"
		case 6:
			block.Type = "h6"
		default:
			block.Type = "heading"
		}
		block.Level = n.Level
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.Paragraph:
		block.Type = "paragraph"
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.List:
		if n.IsOrdered() {
			block.Type = "ordered_list"
		} else {
			block.Type = "unordered_list"
		}
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.ListItem:
		block.Type = "list_item"
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.CodeBlock:
		block.Type = "code_block"
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.FencedCodeBlock:
		block.Type = "fenced_code_block"
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.Blockquote:
		block.Type = "blockquote"
		block.HTML = p.renderNodeToHTML(node, source)
	case *ast.ThematicBreak:
		block.Type = "thematic_break"
		block.HTML = p.renderNodeToHTML(node, source)
	default:
		block.Type = "unknown"
		block.HTML = p.renderNodeToHTML(node, source)
	}

	return block
}

// renderNodeToHTML renders a single AST node to HTML
func (p *MarkdownParser) renderNodeToHTML(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	if err := p.goldmark.Renderer().Render(&buf, source, node); err != nil {
		return ""
	}
	return buf.String()
}

// generateBlockID generates a unique ID for a block based on its content and position
func (p *MarkdownParser) generateBlockID(node ast.Node, source []byte) string {
	var startPos, endPos int
	content := ""
	
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		segment := hasSegment.Segment()
		startPos = segment.Start
		endPos = segment.Stop
		if startPos < len(source) && endPos <= len(source) && endPos > startPos {
			content = string(source[startPos:endPos])
		}
	}
	
	// Create a hash of content + position for uniqueness
	hash := md5.Sum([]byte(fmt.Sprintf("%s-%d-%d", content, startPos, endPos)))
	return fmt.Sprintf("%x", hash)[:8]
}

// DetectNotionSyntax detects Notion-style syntax patterns
func (p *MarkdownParser) DetectNotionSyntax(line string) string {
	trimmed := strings.TrimSpace(line)
	
	// Heading detection
	if strings.HasPrefix(trimmed, "# ") {
		return "h1"
	}
	if strings.HasPrefix(trimmed, "## ") {
		return "h2"
	}
	if strings.HasPrefix(trimmed, "### ") {
		return "h3"
	}
	if strings.HasPrefix(trimmed, "#### ") {
		return "h4"
	}
	if strings.HasPrefix(trimmed, "##### ") {
		return "h5"
	}
	if strings.HasPrefix(trimmed, "###### ") {
		return "h6"
	}
	
	// Checkbox detection (check before list detection)
	if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]") {
		return "checkbox"
	}
	
	// List detection
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
		return "unordered_list"
	}
	
	// Ordered list detection
	if len(trimmed) > 0 && unicode.IsDigit(rune(trimmed[0])) {
		parts := strings.SplitN(trimmed, ". ", 2)
		if len(parts) == 2 {
			return "ordered_list"
		}
	}
	
	// Code block detection
	if strings.HasPrefix(trimmed, "```") {
		return "code_block"
	}
	
	// Blockquote detection
	if strings.HasPrefix(trimmed, "> ") {
		return "blockquote"
	}
	
	return "paragraph"
}