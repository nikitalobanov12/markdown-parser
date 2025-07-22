package tests

import (
	"testing"

	"markdown-parser/internal/parser"
)

func TestMarkdownParser_Parse(t *testing.T) {
	p := parser.NewMarkdownParser()

	tests := []struct {
		name     string
		input    string
		wantHTML bool
	}{
		{
			name:     "empty content",
			input:    "",
			wantHTML: false,
		},
		{
			name:     "heading",
			input:    "# Hello World",
			wantHTML: true,
		},
		{
			name:     "paragraph",
			input:    "This is a paragraph.",
			wantHTML: true,
		},
		{
			name:     "list",
			input:    "- Item 1\n- Item 2",
			wantHTML: true,
		},
		{
			name:     "code block",
			input:    "```go\nfunc main() {}\n```",
			wantHTML: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(tt.input)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			if !result.Success {
				t.Errorf("Parse() success = false, want true")
			}

			if tt.wantHTML && result.HTML == "" {
				t.Errorf("Parse() HTML is empty, expected content")
			}

			if !tt.wantHTML && result.HTML != "" {
				t.Errorf("Parse() HTML = %v, want empty", result.HTML)
			}
		})
	}
}

func TestMarkdownParser_DetectNotionSyntax(t *testing.T) {
	p := parser.NewMarkdownParser()

	tests := []struct {
		input    string
		expected string
	}{
		{"# Heading 1", "h1"},
		{"## Heading 2", "h2"},
		{"### Heading 3", "h3"},
		{"- List item", "unordered_list"},
		{"* List item", "unordered_list"},
		{"1. Ordered item", "ordered_list"},
		{"> Blockquote", "blockquote"},
		{"```javascript", "code_block"},
		{"- [ ] Checkbox", "checkbox"},
		{"- [x] Checked", "checkbox"},
		{"Regular text", "paragraph"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := p.DetectNotionSyntax(tt.input)
			if result != tt.expected {
				t.Errorf("DetectNotionSyntax(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIncrementalParser(t *testing.T) {
	ip := parser.NewIncrementalParser()

	// Test initial parsing
	content1 := "# First heading\n\nSome content"
	result1, err := ip.ParseWithDiff(content1)
	if err != nil {
		t.Errorf("ParseWithDiff() error = %v", err)
		return
	}

	if len(result1.Changes) == 0 {
		t.Errorf("Expected changes for initial parse, got none")
	}

	// Test incremental update
	content2 := "# Updated heading\n\nSome content\n\n## New section"
	result2, err := ip.ParseWithDiff(content2)
	if err != nil {
		t.Errorf("ParseWithDiff() error = %v", err)
		return
	}

	if len(result2.Changes) == 0 {
		t.Errorf("Expected changes for updated content, got none")
	}
}

func TestLineParsing(t *testing.T) {
	ip := parser.NewIncrementalParser()

	tests := []struct {
		line     string
		lineNum  int
		wantType string
	}{
		{"# Test Heading", 1, "h1"},
		{"- List item", 2, "unordered_list"},
		{"Regular paragraph", 3, "paragraph"},
		{"", 4, ""}, // Empty line should return nil
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := ip.ParseLine(tt.line, tt.lineNum)
			
			if tt.wantType == "" {
				if result != nil {
					t.Errorf("ParseLine(%v) = %v, want nil", tt.line, result)
				}
				return
			}

			if result == nil {
				t.Errorf("ParseLine(%v) = nil, want block", tt.line)
				return
			}

			if result.Type != tt.wantType {
				t.Errorf("ParseLine(%v) type = %v, want %v", tt.line, result.Type, tt.wantType)
			}

			if result.Position.Line != tt.lineNum {
				t.Errorf("ParseLine(%v) line = %v, want %v", tt.line, result.Position.Line, tt.lineNum)
			}
		})
	}
}