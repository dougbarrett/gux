package api

import "context"

//go:generate go run gux/cmd/apigen -source=posts.go -output=posts_client_gen.go

// PostsAPI defines the posts endpoints
// @client PostsClient
// @basepath /api/posts
type PostsAPI interface {
	// GetAll returns all posts
	// @route GET /
	GetAll(ctx context.Context) ([]Post, error)

	// GetByID returns a single post by ID
	// @route GET /{id}
	GetByID(ctx context.Context, id int) (*Post, error)

	// Create creates a new post
	// @route POST /
	Create(ctx context.Context, req CreatePostRequest) (*Post, error)

	// Update updates an existing post
	// @route PUT /{id}
	Update(ctx context.Context, id int, req CreatePostRequest) (*Post, error)

	// Delete removes a post
	// @route DELETE /{id}
	Delete(ctx context.Context, id int) error
}
