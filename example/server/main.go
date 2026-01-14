package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gux/example/api"
	"gux/server"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve static files from")
	flag.Parse()

	mux := http.NewServeMux()

	// Create service and wrap with generated HTTP handler
	postsService := NewPostsService()
	postsHandler := api.NewPostsAPIHandler(postsService)

	// Add middleware (logging, CORS, panic recovery)
	postsHandler.Use(
		server.Logger(),
		server.CORS(server.CORSOptions{}),
		server.Recover(),
	)

	postsHandler.RegisterRoutes(mux)

	// WebSocket handler for real-time posts
	postsWSHandler := NewPostsWSHandler(postsService)
	mux.HandleFunc("/ws/posts", postsWSHandler.ServeHTTP)

	// Wire up event callbacks so HTTP API changes broadcast to WebSocket clients
	postsService.SetEventCallbacks(
		func(post api.Post) { postsWSHandler.broadcastEvent("post.created", post) },
		func(post api.Post) { postsWSHandler.broadcastEvent("post.updated", post) },
		func(id int) { postsWSHandler.broadcastEvent("post.deleted", struct{ ID int `json:"id"` }{id}) },
	)

	// SPA handler for static files
	spaHandler := server.NewSPAHandler(*dir)
	mux.HandleFunc("/", spaHandler.ServeHTTP)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Server running at http://localhost%s\n", addr)
	fmt.Printf("Serving static files from: %s\n", *dir)
	fmt.Println("\nAPI endpoints:")
	fmt.Println("  GET    /api/posts      - List all posts")
	fmt.Println("  GET    /api/posts/:id  - Get post by ID")
	fmt.Println("  POST   /api/posts      - Create post")
	fmt.Println("  PUT    /api/posts/:id  - Update post")
	fmt.Println("  DELETE /api/posts/:id  - Delete post")
	fmt.Println("\nWebSocket:")
	fmt.Println("  WS     /ws/posts       - Real-time posts API")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
