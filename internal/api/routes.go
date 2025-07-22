package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"markdown-parser/internal/models"
	"markdown-parser/internal/parser"
)

var markdownParser *parser.MarkdownParser

// SetupRoutes initializes all API routes
func SetupRoutes(r *gin.Engine) {
	markdownParser = parser.NewMarkdownParser()

	api := r.Group("/api")
	{
		api.POST("/parse", parseMarkdown)
		api.POST("/parse-incremental", parseIncremental)
		api.GET("/syntax-check/:syntax", checkSyntax)
	}
}

// parseMarkdown handles bulk markdown parsing
func parseMarkdown(c *gin.Context) {
	var req models.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ParseResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	response, err := markdownParser.Parse(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ParseResponse{
			Success: false,
			Error:   "Failed to parse markdown: " + err.Error(),
		})
		return
	}

	// Include AST if requested
	if req.Format == "ast" {
		// For now, we'll include block information as AST
		response.AST = response.Blocks
	}

	c.JSON(http.StatusOK, response)
}

// parseIncremental handles incremental parsing for real-time updates
func parseIncremental(c *gin.Context) {
	var req models.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ParseResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		return
	}

	response, err := markdownParser.ParseIncremental(req.Content, req.BlockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ParseResponse{
			Success: false,
			Error:   "Failed to parse markdown incrementally: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// checkSyntax checks if a given line matches Notion-style syntax
func checkSyntax(c *gin.Context) {
	syntax := c.Param("syntax")
	if syntax == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Syntax parameter is required",
		})
		return
	}

	detectedType := markdownParser.DetectNotionSyntax(syntax)
	
	c.JSON(http.StatusOK, gin.H{
		"syntax":       syntax,
		"detected_type": detectedType,
		"is_block":     detectedType != "paragraph",
	})
}