package diff

import (
	"crypto/md5"
	"fmt"
	"strings"

	"markdown-parser/internal/models"
)

// BlockDiffer handles block-level diff operations
type BlockDiffer struct {
	previousBlocks map[string]*models.Block
}

// NewBlockDiffer creates a new block differ
func NewBlockDiffer() *BlockDiffer {
	return &BlockDiffer{
		previousBlocks: make(map[string]*models.Block),
	}
}

// ComputeDiff computes the differences between old and new blocks
func (d *BlockDiffer) ComputeDiff(newBlocks map[string]*models.Block) []models.BlockChange {
	var changes []models.BlockChange

	// Track which blocks we've seen in the new version
	seenBlocks := make(map[string]bool)

	// Check for added or modified blocks
	for blockID, newBlock := range newBlocks {
		seenBlocks[blockID] = true
		
		if oldBlock, exists := d.previousBlocks[blockID]; exists {
			// Block exists, check if it's been modified
			if d.hasBlockChanged(oldBlock, newBlock) {
				changes = append(changes, models.BlockChange{
					Type:    "modified",
					BlockID: blockID,
					Block:   newBlock,
				})
			}
		} else {
			// New block added
			changes = append(changes, models.BlockChange{
				Type:    "added",
				BlockID: blockID,
				Block:   newBlock,
			})
		}
	}

	// Check for removed blocks
	for blockID, oldBlock := range d.previousBlocks {
		if !seenBlocks[blockID] {
			changes = append(changes, models.BlockChange{
				Type:    "removed",
				BlockID: blockID,
				Block:   oldBlock,
			})
		}
	}

	// Update the previous blocks for next diff
	d.previousBlocks = d.copyBlocks(newBlocks)

	return changes
}

// hasBlockChanged checks if a block has been modified
func (d *BlockDiffer) hasBlockChanged(oldBlock, newBlock *models.Block) bool {
	// Compare content hashes for efficient change detection
	oldHash := d.computeBlockHash(oldBlock)
	newHash := d.computeBlockHash(newBlock)
	
	return oldHash != newHash
}

// computeBlockHash computes a hash for a block to detect changes
func (d *BlockDiffer) computeBlockHash(block *models.Block) string {
	// Create hash based on block content, type, and level
	content := fmt.Sprintf("%s|%s|%d|%s", block.Type, block.Content, block.Level, block.HTML)
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// copyBlocks creates a deep copy of the blocks map
func (d *BlockDiffer) copyBlocks(blocks map[string]*models.Block) map[string]*models.Block {
	copied := make(map[string]*models.Block)
	for id, block := range blocks {
		copied[id] = d.copyBlock(block)
	}
	return copied
}

// copyBlock creates a deep copy of a single block
func (d *BlockDiffer) copyBlock(block *models.Block) *models.Block {
	if block == nil {
		return nil
	}

	copied := &models.Block{
		ID:       block.ID,
		Type:     block.Type,
		Level:    block.Level,
		Content:  block.Content,
		HTML:     block.HTML,
		Position: block.Position,
	}

	// Copy children if they exist
	if len(block.Children) > 0 {
		copied.Children = make([]*models.Block, len(block.Children))
		for i, child := range block.Children {
			copied.Children[i] = d.copyBlock(child)
		}
	}

	return copied
}

// LineDiffer handles line-by-line diff operations for fine-grained updates
type LineDiffer struct{}

// NewLineDiffer creates a new line differ
func NewLineDiffer() *LineDiffer {
	return &LineDiffer{}
}

// ComputeLineDiff computes line-by-line differences
func (ld *LineDiffer) ComputeLineDiff(oldContent, newContent string) []LineChange {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	return ld.computeLCS(oldLines, newLines)
}

// LineChange represents a change in a line
type LineChange struct {
	Type    string `json:"type"`    // added, removed, unchanged
	LineNum int    `json:"lineNum"`
	Content string `json:"content"`
}

// computeLCS computes the Longest Common Subsequence for diff
func (ld *LineDiffer) computeLCS(oldLines, newLines []string) []LineChange {
	oldLen := len(oldLines)
	newLen := len(newLines)
	
	// Dynamic programming table for LCS
	dp := make([][]int, oldLen+1)
	for i := range dp {
		dp[i] = make([]int, newLen+1)
	}

	// Fill the LCS table
	for i := 1; i <= oldLen; i++ {
		for j := 1; j <= newLen; j++ {
			if oldLines[i-1] == newLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// Backtrack to find changes
	var changes []LineChange
	i, j := oldLen, newLen
	
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			// Lines are the same
			changes = append([]LineChange{{
				Type:    "unchanged",
				LineNum: j,
				Content: newLines[j-1],
			}}, changes...)
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			// Line was added
			changes = append([]LineChange{{
				Type:    "added",
				LineNum: j,
				Content: newLines[j-1],
			}}, changes...)
			j--
		} else if i > 0 {
			// Line was removed
			changes = append([]LineChange{{
				Type:    "removed",
				LineNum: i,
				Content: oldLines[i-1],
			}}, changes...)
			i--
		}
	}

	return changes
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}