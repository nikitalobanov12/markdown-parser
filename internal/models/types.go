package models

import (
	"time"
	"github.com/yuin/goldmark/ast"
)

// ParseRequest represents a request to parse markdown content
type ParseRequest struct {
	Content string `json:"content" binding:"required"`
	BlockID string `json:"blockId,omitempty"`
	Format  string `json:"format,omitempty"` // html, ast, preview
}

// ParseResponse represents the response from parsing
type ParseResponse struct {
	HTML     string            `json:"html"`
	AST      interface{}       `json:"ast,omitempty"`
	Blocks   map[string]*Block `json:"blocks"`
	Changes  []BlockChange     `json:"changes,omitempty"`
	Success  bool              `json:"success"`
	Error    string            `json:"error,omitempty"`
}

// Block represents a parsed markdown block
type Block struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`     // heading, paragraph, list, code_block, etc.
	Level    int       `json:"level"`    // For headings (1-6), list nesting level
	Content  string    `json:"content"`  // Original markdown content
	HTML     string    `json:"html"`     // Rendered HTML
	Position Position  `json:"position"` // Position in source
	Children []*Block  `json:"children,omitempty"`
}

// Position represents the position of content in the source
type Position struct {
	Start int `json:"start"`
	End   int `json:"end"`
	Line  int `json:"line"`
}

// BlockChange represents a change in a block
type BlockChange struct {
	Type    string `json:"type"`    // added, modified, removed
	BlockID string `json:"blockId"`
	Block   *Block `json:"block,omitempty"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`      // parse, subscribe, unsubscribe
	DocumentID string     `json:"documentId,omitempty"`
	Content   string      `json:"content,omitempty"`
	BlockID   string      `json:"blockId,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// WebSocketResponse represents a WebSocket response
type WebSocketResponse struct {
	Type      string      `json:"type"`      // parsed, error, connected
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NotionBlock represents a Notion-style block for real-time updates
type NotionBlock struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Content  map[string]interface{} `json:"content"`
	Children []string               `json:"children,omitempty"`
	Parent   string                 `json:"parent,omitempty"`
}

// ASTNodeInfo represents information about an AST node
type ASTNodeInfo struct {
	Kind     ast.NodeKind `json:"kind"`
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	Level    int          `json:"level,omitempty"`
	Position Position     `json:"position"`
}