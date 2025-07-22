package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"markdown-parser/configs"
	"markdown-parser/internal/api"
	"markdown-parser/internal/websocket"
)

func main() {
	// Set production mode if not already set
	if gin.Mode() != gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Load configuration
	config, err := configs.LoadConfig("configs/config.json")
	if err != nil {
		log.Printf("Error loading config: %v, using defaults", err)
		config = configs.DefaultConfig()
	}

	// Initialize Gin router
	r := gin.Default()

	// Add CORS middleware for React frontend
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, allowedOrigin := range config.Server.AllowOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "markdown-parser",
			"version": "1.0.0",
			"config":  gin.H{
				"max_content_size": config.Parser.MaxContentSize,
				"max_connections":  config.WebSocket.MaxConnections,
			},
		})
	})

	// Initialize API routes
	api.SetupRoutes(r)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		websocket.HandleWebSocket(hub, c)
	})

	// Use Railway's PORT environment variable or fallback to config
	port := os.Getenv("PORT")
	if port == "" {
		port = config.Server.Port
	}

	// Start server
	address := config.Server.Host + ":" + port
	log.Printf("Starting markdown parser service on %s", address)
	log.Printf("CORS origins: %s", strings.Join(config.Server.AllowOrigins, ", "))
	log.Fatal(r.Run(":" + port))
}