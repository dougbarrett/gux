package main

import (
	"context"
	"sync"

	gqapi "github.com/dougbarrett/gux/api"
	"github.com/dougbarrett/gux/example/api"
)

// PostsService implements api.PostsAPI
type PostsService struct {
	mu     sync.RWMutex
	posts  map[int]api.Post
	nextID int

	// Event callbacks for real-time notifications
	onCreated func(api.Post)
	onUpdated func(api.Post)
	onDeleted func(id int)
}

// SetEventCallbacks sets the callbacks for post events
func (s *PostsService) SetEventCallbacks(onCreated func(api.Post), onUpdated func(api.Post), onDeleted func(int)) {
	s.onCreated = onCreated
	s.onUpdated = onUpdated
	s.onDeleted = onDeleted
}

// NewPostsService creates a new PostsService with sample data
func NewPostsService() *PostsService {
	s := &PostsService{
		posts:  make(map[int]api.Post),
		nextID: 1,
	}

	// Add sample data
	samplePosts := []api.Post{
		{ID: 1, UserID: 1, Title: "Hello World", Body: "This is the first post from our Go WASM backend!"},
		{ID: 2, UserID: 1, Title: "Getting Started with Go WASM", Body: "Learn how to build web apps with Go and WebAssembly."},
		{ID: 3, UserID: 2, Title: "API Design Patterns", Body: "Best practices for designing clean APIs in Go."},
	}

	for _, p := range samplePosts {
		s.posts[p.ID] = p
	}
	s.nextID = 4

	return s
}

// GetAll returns all posts
func (s *PostsService) GetAll(ctx context.Context) ([]api.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	posts := make([]api.Post, 0, len(s.posts))
	for _, p := range s.posts {
		posts = append(posts, p)
	}
	return posts, nil
}

// GetByID returns a single post by ID
func (s *PostsService) GetByID(ctx context.Context, id int) (*api.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, ok := s.posts[id]
	if !ok {
		return nil, gqapi.NotFoundf("post %d not found", id)
	}
	return &post, nil
}

// Create creates a new post
func (s *PostsService) Create(ctx context.Context, req api.CreatePostRequest) (*api.Post, error) {
	s.mu.Lock()
	post := api.Post{
		ID:     s.nextID,
		UserID: req.UserID,
		Title:  req.Title,
		Body:   req.Body,
	}
	s.posts[post.ID] = post
	s.nextID++
	s.mu.Unlock()

	// Notify listeners
	if s.onCreated != nil {
		s.onCreated(post)
	}

	return &post, nil
}

// Update updates an existing post
func (s *PostsService) Update(ctx context.Context, id int, req api.CreatePostRequest) (*api.Post, error) {
	s.mu.Lock()
	if _, ok := s.posts[id]; !ok {
		s.mu.Unlock()
		return nil, gqapi.NotFoundf("post %d not found", id)
	}

	post := api.Post{
		ID:     id,
		UserID: req.UserID,
		Title:  req.Title,
		Body:   req.Body,
	}
	s.posts[id] = post
	s.mu.Unlock()

	// Notify listeners
	if s.onUpdated != nil {
		s.onUpdated(post)
	}

	return &post, nil
}

// Delete removes a post
func (s *PostsService) Delete(ctx context.Context, id int) error {
	s.mu.Lock()
	if _, ok := s.posts[id]; !ok {
		s.mu.Unlock()
		return gqapi.NotFoundf("post %d not found", id)
	}

	delete(s.posts, id)
	s.mu.Unlock()

	// Notify listeners
	if s.onDeleted != nil {
		s.onDeleted(id)
	}

	return nil
}
