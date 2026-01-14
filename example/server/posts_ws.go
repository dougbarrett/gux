package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"gux/example/api"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Message represents a typed WebSocket message
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	ID      string          `json:"id,omitempty"`
}

// PostsWSHandler handles WebSocket connections for posts
type PostsWSHandler struct {
	service     *PostsService
	clients     map[*websocket.Conn]bool
	clientsMu   sync.RWMutex
	broadcast   chan Message
}

// NewPostsWSHandler creates a new WebSocket handler
func NewPostsWSHandler(service *PostsService) *PostsWSHandler {
	h := &PostsWSHandler{
		service:   service,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Message, 256),
	}
	go h.runBroadcast()
	return h
}

// ServeHTTP handles WebSocket upgrade and message handling
func (h *PostsWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()

	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, conn)
		h.clientsMu.Unlock()
	}()

	log.Printf("WebSocket client connected")

	// Handle messages
	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(msgData, &msg); err != nil {
			h.sendError(conn, msg.ID, "invalid message format")
			continue
		}

		h.handleMessage(conn, msg)
	}
}

// handleMessage routes messages to appropriate handlers
func (h *PostsWSHandler) handleMessage(conn *websocket.Conn, msg Message) {
	ctx := context.Background()

	switch msg.Type {
	case "posts.list":
		posts, err := h.service.GetAll(ctx)
		if err != nil {
			h.sendError(conn, msg.ID, err.Error())
			return
		}
		h.sendResponse(conn, msg.Type+".response", msg.ID, posts)

	case "posts.get":
		var req struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.sendError(conn, msg.ID, "invalid request payload")
			return
		}
		post, err := h.service.GetByID(ctx, req.ID)
		if err != nil {
			h.sendError(conn, msg.ID, err.Error())
			return
		}
		h.sendResponse(conn, msg.Type+".response", msg.ID, post)

	case "posts.create":
		var req api.CreatePostRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.sendError(conn, msg.ID, "invalid request payload")
			return
		}
		post, err := h.service.Create(ctx, req)
		if err != nil {
			h.sendError(conn, msg.ID, err.Error())
			return
		}
		h.sendResponse(conn, msg.Type+".response", msg.ID, post)
		// Broadcast to all clients
		h.broadcastEvent("post.created", post)

	case "posts.update":
		var req struct {
			ID     int                   `json:"id"`
			Update api.CreatePostRequest `json:"update"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.sendError(conn, msg.ID, "invalid request payload")
			return
		}
		post, err := h.service.Update(ctx, req.ID, req.Update)
		if err != nil {
			h.sendError(conn, msg.ID, err.Error())
			return
		}
		h.sendResponse(conn, msg.Type+".response", msg.ID, post)
		// Broadcast to all clients
		h.broadcastEvent("post.updated", post)

	case "posts.delete":
		var req struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.sendError(conn, msg.ID, "invalid request payload")
			return
		}
		if err := h.service.Delete(ctx, req.ID); err != nil {
			h.sendError(conn, msg.ID, err.Error())
			return
		}
		h.sendResponse(conn, msg.Type+".response", msg.ID, struct{ Success bool }{true})
		// Broadcast to all clients
		h.broadcastEvent("post.deleted", struct{ ID int `json:"id"` }{req.ID})

	case "posts.subscribe":
		// Client is subscribed by default when connected
		h.sendResponse(conn, msg.Type+".response", msg.ID, struct{ Subscribed bool }{true})

	case "posts.unsubscribe":
		h.sendResponse(conn, msg.Type+".response", msg.ID, struct{ Subscribed bool }{false})

	default:
		h.sendError(conn, msg.ID, "unknown message type: "+msg.Type)
	}
}

// sendResponse sends a typed response message
func (h *PostsWSHandler) sendResponse(conn *websocket.Conn, msgType, id string, payload any) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	// The response needs to include the ID in the payload for correlation
	resp := struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
		ID      string          `json:"id,omitempty"`
	}{
		Type:    msgType,
		Payload: payloadBytes,
		ID:      id,
	}

	if err := conn.WriteJSON(resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

// sendError sends an error message
func (h *PostsWSHandler) sendError(conn *websocket.Conn, id, message string) {
	h.sendResponse(conn, "error", id, struct {
		Message string `json:"message"`
	}{message})
}

// broadcastEvent sends an event to all connected clients
func (h *PostsWSHandler) broadcastEvent(eventType string, payload any) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal broadcast: %v", err)
		return
	}

	h.broadcast <- Message{
		Type:    eventType,
		Payload: payloadBytes,
	}
}

// runBroadcast handles broadcasting messages to all clients
func (h *PostsWSHandler) runBroadcast() {
	for msg := range h.broadcast {
		h.clientsMu.RLock()
		for conn := range h.clients {
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("Broadcast error: %v", err)
			}
		}
		h.clientsMu.RUnlock()
	}
}
