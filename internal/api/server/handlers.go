package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"goquery/internal/api"
)

// PostsHandler implements the PostsAPI as HTTP handlers
type PostsHandler struct {
	mu     sync.RWMutex
	posts  map[int]api.Post
	nextID int
}

// NewPostsHandler creates a new PostsHandler with sample data
func NewPostsHandler() *PostsHandler {
	h := &PostsHandler{
		posts:  make(map[int]api.Post),
		nextID: 1,
	}

	// Add some sample data
	samplePosts := []api.Post{
		{ID: 1, UserID: 1, Title: "Hello World", Body: "This is the first post from our Go WASM backend!"},
		{ID: 2, UserID: 1, Title: "Getting Started with Go WASM", Body: "Learn how to build web apps with Go and WebAssembly."},
		{ID: 3, UserID: 2, Title: "API Design Patterns", Body: "Best practices for designing clean APIs in Go."},
	}

	for _, p := range samplePosts {
		h.posts[p.ID] = p
	}
	h.nextID = 4

	return h
}

// RegisterRoutes adds the posts routes to the given mux
func (h *PostsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/posts", h.handlePosts)
	mux.HandleFunc("/api/posts/", h.handlePostByID)
}

func (h *PostsHandler) handlePosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAll(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PostsHandler) handlePostByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/posts/123
	path := strings.TrimPrefix(r.URL.Path, "/api/posts/")

	// Handle trailing slash case: /api/posts/ -> delegate to handlePosts
	if path == "" {
		h.handlePosts(w, r)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getByID(w, r, id)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PostsHandler) getAll(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	posts := make([]api.Post, 0, len(h.posts))
	for _, p := range h.posts {
		posts = append(posts, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostsHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	post, ok := h.posts[id]
	if !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostsHandler) create(w http.ResponseWriter, r *http.Request) {
	var req api.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	post := api.Post{
		ID:     h.nextID,
		UserID: req.UserID,
		Title:  req.Title,
		Body:   req.Body,
	}
	h.posts[post.ID] = post
	h.nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func (h *PostsHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var req api.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.posts[id]; !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	post := api.Post{
		ID:     id,
		UserID: req.UserID,
		Title:  req.Title,
		Body:   req.Body,
	}
	h.posts[id] = post

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostsHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.posts[id]; !ok {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	delete(h.posts, id)
	w.WriteHeader(http.StatusNoContent)
}
