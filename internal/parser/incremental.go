package parser

import (
	"crypto/md5"
	"fmt"
	"strings"

	"markdown-parser/internal/models"
	"markdown-parser/pkg/diff"
)

// IncrementalParser handles real-time parsing with diff detection
type IncrementalParser struct {
	baseParser *MarkdownParser
	differ     *diff.BlockDiffer
	lineDiffer *diff.LineDiffer
}

// NewIncrementalParser creates a new incremental parser
func NewIncrementalParser() *IncrementalParser {
	return &IncrementalParser{
		baseParser: NewMarkdownParser(),
		differ:     diff.NewBlockDiffer(),
		lineDiffer: diff.NewLineDiffer(),
	}
}

// ParseWithDiff parses content and returns changes from previous version
func (ip *IncrementalParser) ParseWithDiff(content string) (*models.ParseResponse, error) {
	// Parse the full content
	result, err := ip.baseParser.Parse(content)
	if err != nil {
		return nil, err
	}

	// Compute block-level differences
	changes := ip.differ.ComputeDiff(result.Blocks)
	result.Changes = changes

	return result, nil
}

// ParseLine parses a single line and detects Notion-style syntax
func (ip *IncrementalParser) ParseLine(line string, lineNumber int) *models.Block {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}

	// Detect syntax type first
	syntaxType := ip.baseParser.DetectNotionSyntax(line)

	// Create a basic block with the detected type
	block := &models.Block{
		ID:      generateLineID(line, lineNumber),
		Type:    syntaxType,
		Content: line,
		HTML:    ip.renderLineToHTML(line, syntaxType),
		Position: models.Position{
			Line:  lineNumber,
			Start: 0,
			End:   len(line),
		},
	}

	// Try to parse with goldmark and update if we get better results
	result, err := ip.baseParser.Parse(line)
	if err == nil && len(result.Blocks) > 0 {
		// Use goldmark's result but keep our detected type if it's more specific
		for _, goldmarkBlock := range result.Blocks {
			if goldmarkBlock.Type != "unknown" && 
			   goldmarkBlock.HTML != "" && 
			   (syntaxType == "paragraph" || syntaxType == goldmarkBlock.Type) {
				block.HTML = goldmarkBlock.HTML
				if syntaxType == "paragraph" {
					block.Type = goldmarkBlock.Type
				}
				break
			}
		}
	}

	return block
}

// renderLineToHTML renders a single line to HTML based on syntax type
func (ip *IncrementalParser) renderLineToHTML(line, syntaxType string) string {
	trimmed := strings.TrimSpace(line)

	switch syntaxType {
	case "h1":
		content := strings.TrimPrefix(trimmed, "# ")
		return "<h1>" + content + "</h1>"
	case "h2":
		content := strings.TrimPrefix(trimmed, "## ")
		return "<h2>" + content + "</h2>"
	case "h3":
		content := strings.TrimPrefix(trimmed, "### ")
		return "<h3>" + content + "</h3>"
	case "h4":
		content := strings.TrimPrefix(trimmed, "#### ")
		return "<h4>" + content + "</h4>"
	case "h5":
		content := strings.TrimPrefix(trimmed, "##### ")
		return "<h5>" + content + "</h5>"
	case "h6":
		content := strings.TrimPrefix(trimmed, "###### ")
		return "<h6>" + content + "</h6>"
	case "unordered_list":
		content := strings.TrimPrefix(trimmed, "- ")
		content = strings.TrimPrefix(content, "* ")
		content = strings.TrimPrefix(content, "+ ")
		return "<ul><li>" + content + "</li></ul>"
	case "ordered_list":
		parts := strings.SplitN(trimmed, ". ", 2)
		if len(parts) == 2 {
			return "<ol><li>" + parts[1] + "</li></ol>"
		}
		return "<p>" + line + "</p>"
	case "blockquote":
		content := strings.TrimPrefix(trimmed, "> ")
		return "<blockquote><p>" + content + "</p></blockquote>"
	case "code_block":
		if strings.HasPrefix(trimmed, "```") {
			lang := strings.TrimPrefix(trimmed, "```")
			if lang == "" {
				return "<pre><code>"
			}
			return "<pre><code class=\"language-" + lang + "\">"
		}
		return "<pre><code>" + line + "</code></pre>"
	case "checkbox":
		if strings.Contains(trimmed, "- [x]") {
			content := strings.Replace(trimmed, "- [x]", "", 1)
			return "<ul><li><input type=\"checkbox\" checked disabled>" + strings.TrimSpace(content) + "</li></ul>"
		}
		content := strings.Replace(trimmed, "- [ ]", "", 1)
		return "<ul><li><input type=\"checkbox\" disabled>" + strings.TrimSpace(content) + "</li></ul>"
	default:
		return "<p>" + line + "</p>"
	}
}

// generateLineID generates a unique ID for a line
func generateLineID(line string, lineNumber int) string {
	content := strings.TrimSpace(line)
	if content == "" {
		content = "empty"
	}
	return fmt.Sprintf("line_%d_%x", lineNumber, md5.Sum([]byte(content)))[:12]
}