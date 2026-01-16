# API Generation

Gux's code generator (`apigen`) creates type-safe HTTP clients and server handlers from Go interface definitions. This eliminates boilerplate while ensuring compile-time type safety.

## Overview

Define your API once as a Go interface:

```go
// @client PostsClient
// @basepath /api/posts
type PostsAPI interface {
    // @route GET /
    GetAll(ctx context.Context) ([]Post, error)

    // @route POST /
    Create(ctx context.Context, req CreatePostRequest) (*Post, error)
}
```

Generate both client and server code:

```bash
go run gux/cmd/apigen -source=posts.go -output=posts_client_gen.go
```

This generates:
- **Client** (`posts_client_gen.go`) — Type-safe HTTP client for WASM
- **Server** (`posts_server_gen.go`) — HTTP handler wrapper

## Annotations

### @client

Names the generated client struct.

```go
// @client PostsClient
type PostsAPI interface { ... }
```

Generated:
```go
type PostsClient struct { ... }

func NewPostsClient(opts ...ClientOption) *PostsClient { ... }
```

### @basepath

Sets the base URL path for all endpoints.

```go
// @basepath /api/v1/posts
```

All routes will be prefixed with this path:
- `GET /` becomes `GET /api/v1/posts/`
- `GET /{id}` becomes `GET /api/v1/posts/{id}`

### @route

Defines the HTTP method and path for each interface method.

```go
// @route GET /
// @route POST /
// @route GET /{id}
// @route PUT /{id}
// @route DELETE /{id}
// @route PATCH /{id}/status
```

## Path Parameters

Path parameters use `{name}` syntax and are automatically extracted from method arguments.

```go
// @route GET /{id}
GetByID(ctx context.Context, id int) (*Post, error)

// @route GET /{userId}/posts/{postId}
GetUserPost(ctx context.Context, userId int, postId int) (*Post, error)
```

**Rules:**
- Parameter names must match function argument names exactly
- Parameters must be `int` or `string` types
- Order in the path determines URL structure

## Request Bodies

The generator automatically detects request body parameters:

```go
// @route POST /
Create(ctx context.Context, req CreatePostRequest) (*Post, error)
```

**Detection rules:**
1. Struct types (not primitives) are treated as request bodies
2. The `context.Context` parameter is always skipped
3. Path parameters are extracted, remaining structs become the body

### Example with path and body

```go
// @route PUT /{id}
Update(ctx context.Context, id int, req UpdateRequest) (*Post, error)
```

Generated client code:
```go
func (c *PostsClient) Update(id int, req UpdateRequest) (*Post, error) {
    path := fmt.Sprintf("/api/posts/%d", id)
    body, _ := json.Marshal(req)
    resp, err := fetch.Put(c.baseURL+path, string(body), c.headers)
    // ...
}
```

## Return Types

The generator handles various return type patterns:

```go
// Single value + error
GetAll(ctx context.Context) ([]Post, error)

// Pointer + error
GetByID(ctx context.Context, id int) (*Post, error)

// Error only (for DELETE)
Delete(ctx context.Context, id int) error
```

## Generated Client

### Constructor

```go
client := api.NewPostsClient()

// With options
client := api.NewPostsClient(
    api.WithBaseURL("https://api.example.com"),
    api.WithHeader("Authorization", "Bearer token"),
)
```

### Client Options

```go
// Set base URL
api.WithBaseURL(url string)

// Set base path (appended to base URL)
api.WithBasePath(path string)

// Add custom header
api.WithHeader(key, value string)
```

### Method Calls

```go
// No parameters
posts, err := client.GetAll()

// Path parameter
post, err := client.GetByID(123)

// Request body
created, err := client.Create(api.CreatePostRequest{
    Title: "Hello",
    Body:  "World",
})

// Path parameter + body
updated, err := client.Update(123, api.UpdateRequest{
    Title: "Updated",
})

// Error only return
err := client.Delete(123)
```

## Generated Server Handler

### Handler Struct

```go
type PostsAPIHandler struct {
    service     PostsAPI
    middlewares []server.Middleware
}

func NewPostsAPIHandler(service PostsAPI) *PostsAPIHandler {
    return &PostsAPIHandler{service: service}
}
```

### Adding Middleware

```go
handler := api.NewPostsAPIHandler(service)
handler.Use(
    server.Logger(),
    server.CORS(server.CORSOptions{}),
    server.Recover(),
)
```

### Registering Routes

```go
mux := http.NewServeMux()
handler.RegisterRoutes(mux)
```

This registers all routes defined in the interface:
- `GET /api/posts/` → `GetAll`
- `GET /api/posts/{id}` → `GetByID`
- `POST /api/posts/` → `Create`
- etc.

### Implementing the Service

Your service must implement the interface:

```go
type PostsService struct {
    // your fields
}

func (s *PostsService) GetAll(ctx context.Context) ([]Post, error) {
    // implementation
}

func (s *PostsService) GetByID(ctx context.Context, id int) (*Post, error) {
    // implementation
}

// ... implement all methods
```

## Error Handling

### Client-Side

The generated client returns errors for:
- Network failures
- Non-2xx status codes
- JSON parsing errors

```go
post, err := client.GetByID(999)
if err != nil {
    // Handle error
    components.Toast(err.Error(), components.ToastError)
}
```

### Server-Side

Use the `api` package for structured errors:

```go
import gqapi "github.com/dougbarrett/guxapi"

func (s *PostsService) GetByID(ctx context.Context, id int) (*Post, error) {
    post, ok := s.posts[id]
    if !ok {
        return nil, gqapi.NotFoundf("post %d not found", id)
    }
    return &post, nil
}
```

Available error constructors:
- `api.NotFound(message)` — 404
- `api.BadRequest(message)` — 400
- `api.Unauthorized(message)` — 401
- `api.Forbidden(message)` — 403
- `api.Conflict(message)` — 409
- `api.InternalError(message)` — 500

Format variants: `NotFoundf`, `BadRequestf`, etc.

## Complete Example

### Interface Definition

```go
// api/users.go
package api

import "context"

//go:generate go run gux/cmd/apigen -source=users.go -output=users_client_gen.go

// @client UsersClient
// @basepath /api/users
type UsersAPI interface {
    // @route GET /
    List(ctx context.Context) ([]User, error)

    // @route GET /{id}
    Get(ctx context.Context, id int) (*User, error)

    // @route POST /
    Create(ctx context.Context, req CreateUserRequest) (*User, error)

    // @route PUT /{id}
    Update(ctx context.Context, id int, req UpdateUserRequest) (*User, error)

    // @route DELETE /{id}
    Delete(ctx context.Context, id int) error

    // @route POST /{id}/avatar
    UploadAvatar(ctx context.Context, id int, req AvatarRequest) (*User, error)

    // @route GET /{userId}/posts
    GetUserPosts(ctx context.Context, userId int) ([]Post, error)
}
```

### Types

```go
// api/types.go
package api

type User struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    AvatarURL string `json:"avatarUrl,omitempty"`
}

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UpdateUserRequest struct {
    Name  string `json:"name,omitempty"`
    Email string `json:"email,omitempty"`
}

type AvatarRequest struct {
    URL string `json:"url"`
}
```

### Generated Client Usage

```go
// WASM frontend
client := api.NewUsersClient()

// List all users
users, err := client.List()

// Get single user
user, err := client.Get(123)

// Create user
newUser, err := client.Create(api.CreateUserRequest{
    Name:  "John",
    Email: "john@example.com",
})

// Update user
updated, err := client.Update(123, api.UpdateUserRequest{
    Name: "John Doe",
})

// Delete user
err := client.Delete(123)

// Nested route
posts, err := client.GetUserPosts(123)
```

### Server Implementation

```go
// server/users.go
type UsersService struct {
    db *Database
}

func (s *UsersService) List(ctx context.Context) ([]api.User, error) {
    return s.db.GetAllUsers()
}

func (s *UsersService) Get(ctx context.Context, id int) (*api.User, error) {
    user, err := s.db.GetUser(id)
    if err != nil {
        return nil, gqapi.NotFoundf("user %d not found", id)
    }
    return user, nil
}

// ... implement remaining methods

// server/main.go
func main() {
    service := &UsersService{db: NewDatabase()}
    handler := api.NewUsersAPIHandler(service)
    handler.Use(server.Logger(), server.CORS(server.CORSOptions{}))
    handler.RegisterRoutes(mux)
}
```

## Best Practices

1. **Keep interfaces focused** — One interface per resource type
2. **Use consistent naming** — `GetAll`, `GetByID`, `Create`, `Update`, `Delete`
3. **Return pointers for single items** — `*Post` not `Post`
4. **Return slices for collections** — `[]Post` not `*[]Post`
5. **Always include context** — Even if unused, for future middleware
6. **Use meaningful error messages** — Include IDs and context
7. **Group related types** — Keep request/response types near the interface
