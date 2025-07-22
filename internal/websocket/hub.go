package websocket

import (
	"encoding/json"
	"log"
	"time"

	"markdown-parser/internal/models"
	"markdown-parser/internal/parser"
)

// Hub maintains active WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	parser     *parser.MarkdownParser
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		parser:     parser.NewMarkdownParser(),
	}
}

// Run starts the hub event loop
func (h *Hub) Run() {
	log.Println("WebSocket hub started")
	
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))
			
			// Send connection confirmation
			response := models.WebSocketResponse{
				Type:      "connected",
				Success:   true,
				Timestamp: time.Now(),
			}
			
			if data, err := json.Marshal(response); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			// Broadcast message to all connected clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// HandleMessage processes incoming WebSocket messages
func (h *Hub) HandleMessage(client *Client, messageData []byte) {
	var msg models.WebSocketMessage
	if err := json.Unmarshal(messageData, &msg); err != nil {
		h.sendError(client, "Invalid message format: "+err.Error())
		return
	}

	switch msg.Type {
	case "parse":
		h.handleParse(client, msg)
	case "parse_incremental":
		h.handleParseIncremental(client, msg)
	case "subscribe":
		h.handleSubscribe(client, msg)
	case "unsubscribe":
		h.handleUnsubscribe(client, msg)
	default:
		h.sendError(client, "Unknown message type: "+msg.Type)
	}
}

// handleParse processes markdown parsing requests
func (h *Hub) handleParse(client *Client, msg models.WebSocketMessage) {
	if msg.Content == "" {
		h.sendError(client, "Content is required for parsing")
		return
	}

	// Parse markdown
	result, err := h.parser.Parse(msg.Content)
	if err != nil {
		h.sendError(client, "Failed to parse markdown: "+err.Error())
		return
	}

	// Send response
	response := models.WebSocketResponse{
		Type:      "parsed",
		Success:   true,
		Data:      result,
		Timestamp: time.Now(),
	}

	h.sendToClient(client, response)
}

// handleParseIncremental processes incremental parsing requests
func (h *Hub) handleParseIncremental(client *Client, msg models.WebSocketMessage) {
	if msg.Content == "" {
		h.sendError(client, "Content is required for incremental parsing")
		return
	}

	// Parse markdown incrementally
	result, err := h.parser.ParseIncremental(msg.Content, msg.BlockID)
	if err != nil {
		h.sendError(client, "Failed to parse markdown incrementally: "+err.Error())
		return
	}

	// Send response
	response := models.WebSocketResponse{
		Type:      "parsed_incremental",
		Success:   true,
		Data:      result,
		Timestamp: time.Now(),
	}

	h.sendToClient(client, response)
	
	// Also broadcast to other clients subscribed to the same document
	if msg.DocumentID != "" {
		h.broadcastToDocument(msg.DocumentID, response)
	}
}

// handleSubscribe handles document subscription requests
func (h *Hub) handleSubscribe(client *Client, msg models.WebSocketMessage) {
	if msg.DocumentID == "" {
		h.sendError(client, "Document ID is required for subscription")
		return
	}

	// Add client to document subscription
	client.subscribedDocuments[msg.DocumentID] = true
	
	response := models.WebSocketResponse{
		Type:      "subscribed",
		Success:   true,
		Data:      map[string]string{"documentId": msg.DocumentID},
		Timestamp: time.Now(),
	}

	h.sendToClient(client, response)
}

// handleUnsubscribe handles document unsubscription requests
func (h *Hub) handleUnsubscribe(client *Client, msg models.WebSocketMessage) {
	if msg.DocumentID == "" {
		h.sendError(client, "Document ID is required for unsubscription")
		return
	}

	// Remove client from document subscription
	delete(client.subscribedDocuments, msg.DocumentID)
	
	response := models.WebSocketResponse{
		Type:      "unsubscribed",
		Success:   true,
		Data:      map[string]string{"documentId": msg.DocumentID},
		Timestamp: time.Now(),
	}

	h.sendToClient(client, response)
}

// sendError sends an error response to a client
func (h *Hub) sendError(client *Client, errorMsg string) {
	response := models.WebSocketResponse{
		Type:      "error",
		Success:   false,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}

	h.sendToClient(client, response)
}

// sendToClient sends a response to a specific client
func (h *Hub) sendToClient(client *Client, response models.WebSocketResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

// broadcastToDocument broadcasts a message to all clients subscribed to a document
func (h *Hub) broadcastToDocument(documentID string, response models.WebSocketResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling broadcast response: %v", err)
		return
	}

	for client := range h.clients {
		if client.subscribedDocuments[documentID] {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
}