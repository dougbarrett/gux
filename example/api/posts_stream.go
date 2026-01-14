//go:build js && wasm

package api

import (
	"gux/ws"
)

// PostEvent represents a real-time event from the posts WebSocket
type PostEvent struct {
	Type string // "created", "updated", "deleted"
	Post *Post  // The post data (available for created/updated)
	ID   int    // The post ID (available for all events, especially deleted)
}

// Subscription manages a WebSocket connection for real-time updates
type Subscription struct {
	client *ws.Client
}

// Subscribe connects to the WebSocket and calls handler for each event.
// Usage:
//
//	sub, err := posts.Subscribe(func(event api.PostEvent) {
//	    switch event.Type {
//	    case "created":
//	        fmt.Println("New post:", event.Post.Title)
//	    case "updated":
//	        fmt.Println("Updated:", event.Post.Title)
//	    case "deleted":
//	        fmt.Println("Deleted post ID:", event.ID)
//	    }
//	})
func (c *PostsClient) Subscribe(handler func(PostEvent)) (*Subscription, error) {
	// Build WebSocket URL from HTTP base URL
	wsURL := "ws://localhost:8093/ws/posts"
	if c.cfg.baseURL != "" {
		wsURL = c.cfg.baseURL
		if len(wsURL) > 4 && wsURL[:4] == "http" {
			wsURL = "ws" + wsURL[4:]
		}
		wsURL += "/ws/posts"
	}

	sub := &Subscription{}
	sub.client = ws.NewClient(wsURL)

	if err := sub.client.Connect(); err != nil {
		return nil, err
	}

	// Register handlers for each event type
	ws.OnTyped(sub.client, "post.created", func(post Post) {
		handler(PostEvent{Type: "created", Post: &post, ID: post.ID})
	})

	ws.OnTyped(sub.client, "post.updated", func(post Post) {
		handler(PostEvent{Type: "updated", Post: &post, ID: post.ID})
	})

	ws.OnTyped(sub.client, "post.deleted", func(data struct{ ID int `json:"id"` }) {
		handler(PostEvent{Type: "deleted", ID: data.ID})
	})

	// Subscribe to events on the server
	sub.client.Send("posts.subscribe", struct{}{})

	return sub, nil
}

// Close closes the subscription
func (s *Subscription) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// IsConnected returns true if connected
func (s *Subscription) IsConnected() bool {
	return s.client != nil && s.client.IsConnected()
}
