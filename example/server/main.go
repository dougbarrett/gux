package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"goquery/example/api"
	"goquery/server"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")
	dir := flag.String("dir", ".", "Directory to serve static files from")
	flag.Parse()

	mux := http.NewServeMux()

	// Create service and wrap with generated HTTP handler
	postsService := NewPostsService()
	postsHandler := api.NewPostsAPIHandler(postsService)
	postsHandler.RegisterRoutes(mux)

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

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
