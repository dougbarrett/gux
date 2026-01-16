# WebSocket

Gux provides two levels of WebSocket support:

1. **High-level API** — Type-safe subscriptions that mirror HTTP client patterns
2. **Low-level Client** — Full control with typed message handlers

## High-Level: Type-Safe Subscriptions

The recommended approach mirrors your HTTP API client pattern:

### Defining Events

```go
// api/posts_stream.go
//go:build js && wasm

package api

type PostEvent struct {
    Type string // "created", "updated", "deleted"
    Post *Post  // Available for created/updated
    ID   int    // Available for all events
}
```

### Subscribing to Events

```go
// Subscribe mirrors the HTTP client pattern
sub, err := posts.Subscribe(func(event api.PostEvent) {
    switch event.Type {
    case "created":
        fmt.Println("New post:", event.Post.Title)
        addPostToUI(event.Post)

    case "updated":
        fmt.Println("Updated:", event.Post.Title)
        updatePostInUI(event.Post)

    case "deleted":
        fmt.Println("Deleted ID:", event.ID)
        removePostFromUI(event.ID)
    }
})
if err != nil {
    components.Toast("Connection failed", components.ToastError)
    return
}
defer sub.Close()
```

### Implementing Subscribe

```go
// api/posts_stream.go
func (c *PostsClient) Subscribe(handler func(PostEvent)) (*Subscription, error) {
    // Build WebSocket URL from HTTP client's base URL
    wsURL := c.cfg.baseURL
    if len(wsURL) > 4 && wsURL[:4] == "http" {
        wsURL = "ws" + wsURL[4:]
    }
    wsURL += "/ws/posts"

    sub := &Subscription{}
    sub.client = ws.NewClient(wsURL)

    if err := sub.client.Connect(); err != nil {
        return nil, err
    }

    // Register typed handlers
    ws.OnTyped(sub.client, "post.created", func(post Post) {
        handler(PostEvent{Type: "created", Post: &post, ID: post.ID})
    })

    ws.OnTyped(sub.client, "post.updated", func(post Post) {
        handler(PostEvent{Type: "updated", Post: &post, ID: post.ID})
    })

    ws.OnTyped(sub.client, "post.deleted", func(data struct{ ID int `json:"id"` }) {
        handler(PostEvent{Type: "deleted", ID: data.ID})
    })

    // Subscribe to events on server
    sub.client.Send("posts.subscribe", struct{}{})

    return sub, nil
}

type Subscription struct {
    client *ws.Client
}

func (s *Subscription) Close() error {
    if s.client != nil {
        return s.client.Close()
    }
    return nil
}

func (s *Subscription) IsConnected() bool {
    return s.client != nil && s.client.IsConnected()
}
```

## Low-Level: WebSocket Client

For full control over WebSocket communication:

### Creating a Client

```go
import "github.com/dougbarrett/gux/ws"

client := ws.NewClient("ws://localhost:8080/ws",
    ws.WithOnOpen(func() {
        fmt.Println("Connected!")
    }),
    ws.WithOnClose(func(code int, reason string) {
        fmt.Printf("Closed: %d - %s\n", code, reason)
    }),
    ws.WithOnError(func(err error) {
        fmt.Println("Error:", err)
    }),
    ws.WithOnMessage(func(msg ws.Message) {
        fmt.Println("Raw message:", msg.Type, msg.Payload)
    }),
)
```

### Connecting

```go
err := client.Connect()
if err != nil {
    log.Fatal("Connection failed:", err)
}
defer client.Close()

// Check connection state
if client.IsConnected() {
    // Ready to send/receive
}

// Connection states
state := client.State()
// ws.StateConnecting
// ws.StateOpen
// ws.StateClosing
// ws.StateClosed
```

### Sending Messages

```go
// Typed message (recommended)
client.Send("chat.message", ChatMessage{
    Room: "general",
    Text: "Hello everyone!",
})

// The message is wrapped as:
// {"type": "chat.message", "payload": {"room": "general", "text": "Hello everyone!"}}

// Raw string (for special cases)
client.SendRaw(`{"custom": "format"}`)
```

### Receiving Messages

```go
// Register typed handler
ws.OnTyped(client, "chat.message", func(msg ChatMessage) {
    fmt.Printf("[%s] %s: %s\n", msg.Room, msg.Author, msg.Text)
})

// Register multiple handlers
ws.OnTyped(client, "user.joined", func(event UserEvent) {
    fmt.Println("User joined:", event.Username)
})

ws.OnTyped(client, "user.left", func(event UserEvent) {
    fmt.Println("User left:", event.Username)
})

// Raw handler (receives json.RawMessage)
client.On("system.notice", func(payload json.RawMessage) {
    var notice SystemNotice
    json.Unmarshal(payload, &notice)
    showNotification(notice.Message)
})
```

### Request/Response Pattern

For RPC-style communication:

```go
// Make a request and wait for response
response, err := client.Request("room.join", JoinRequest{
    Room: "general",
})
if err != nil {
    log.Fatal(err)
}

var result JoinResponse
json.Unmarshal(response, &result)
fmt.Println("Joined:", result.Room)

// Type-safe version
resp, err := ws.RequestTyped[JoinRequest, JoinResponse](
    client,
    "room.join",
    JoinRequest{Room: "general"},
)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Members:", resp.MemberCount)
```

## Message Protocol

Gux WebSocket uses a simple JSON message format:

```json
{
    "type": "message.type",
    "payload": { ... },
    "id": "optional-correlation-id"
}
```

- **type** — Message type for routing
- **payload** — Arbitrary JSON data
- **id** — Used for request/response correlation

## Server-Side Implementation

### WebSocket Handler

```go
// server/posts_ws.go
package main

import (
    "encoding/json"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
    "yourapp/api"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type PostsWSHandler struct {
    service *PostsService
    clients map[*websocket.Conn]bool
    mu      sync.RWMutex
}

func NewPostsWSHandler(service *PostsService) *PostsWSHandler {
    return &PostsWSHandler{
        service: service,
        clients: make(map[*websocket.Conn]bool),
    }
}

func (h *PostsWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    // Register client
    h.mu.Lock()
    h.clients[conn] = true
    h.mu.Unlock()

    defer func() {
        h.mu.Lock()
        delete(h.clients, conn)
        h.mu.Unlock()
    }()

    // Handle incoming messages
    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }

        var msg struct {
            Type    string          `json:"type"`
            Payload json.RawMessage `json:"payload"`
            ID      string          `json:"id"`
        }
        if err := json.Unmarshal(message, &msg); err != nil {
            continue
        }

        h.handleMessage(conn, msg.Type, msg.Payload, msg.ID)
    }
}

func (h *PostsWSHandler) handleMessage(conn *websocket.Conn, msgType string, payload json.RawMessage, id string) {
    switch msgType {
    case "posts.subscribe":
        // Client wants to receive post events
        // Already registered by connecting

    case "posts.getAll":
        // Request/response example
        posts, _ := h.service.GetAll(context.Background())
        h.sendResponse(conn, "posts.getAll", posts, id)
    }
}

func (h *PostsWSHandler) sendResponse(conn *websocket.Conn, msgType string, payload any, id string) {
    data, _ := json.Marshal(map[string]any{
        "type":    msgType,
        "payload": payload,
        "id":      id,
    })
    conn.WriteMessage(websocket.TextMessage, data)
}

// Broadcast to all connected clients
func (h *PostsWSHandler) broadcastEvent(eventType string, payload any) {
    data, _ := json.Marshal(map[string]any{
        "type":    eventType,
        "payload": payload,
    })

    h.mu.RLock()
    defer h.mu.RUnlock()

    for conn := range h.clients {
        conn.WriteMessage(websocket.TextMessage, data)
    }
}
```

### Wiring Events

Connect your HTTP API changes to WebSocket broadcasts:

```go
// server/main.go
func main() {
    mux := http.NewServeMux()

    // Create service
    service := NewPostsService()

    // HTTP API handler
    httpHandler := api.NewPostsAPIHandler(service)
    httpHandler.Use(server.Logger(), server.CORS(server.CORSOptions{}))
    httpHandler.RegisterRoutes(mux)

    // WebSocket handler
    wsHandler := NewPostsWSHandler(service)
    mux.HandleFunc("/ws/posts", wsHandler.ServeHTTP)

    // Wire up event callbacks
    service.SetEventCallbacks(
        func(post api.Post) {
            wsHandler.broadcastEvent("post.created", post)
        },
        func(post api.Post) {
            wsHandler.broadcastEvent("post.updated", post)
        },
        func(id int) {
            wsHandler.broadcastEvent("post.deleted", struct {
                ID int `json:"id"`
            }{id})
        },
    )

    http.ListenAndServe(":8080", mux)
}
```

### Service with Callbacks

```go
// server/posts.go
type PostsService struct {
    mu     sync.RWMutex
    posts  map[int]api.Post
    nextID int

    // Event callbacks
    onCreated func(api.Post)
    onUpdated func(api.Post)
    onDeleted func(id int)
}

func (s *PostsService) SetEventCallbacks(
    onCreated func(api.Post),
    onUpdated func(api.Post),
    onDeleted func(int),
) {
    s.onCreated = onCreated
    s.onUpdated = onUpdated
    s.onDeleted = onDeleted
}

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

    // Notify WebSocket clients
    if s.onCreated != nil {
        s.onCreated(post)
    }

    return &post, nil
}
```

## State Integration

Use `WebSocketStore` for integrated state management:

```go
import "github.com/dougbarrett/gux/state"

wsStore := state.NewWebSocketStore(state.WebSocketConfig{
    URL: "ws://localhost:8080/ws",
    OnOpen: func() {
        components.Toast("Connected!", components.ToastSuccess)
    },
    OnClose: func(code int, reason string) {
        components.Toast("Disconnected", components.ToastWarning)
    },
    OnError: func(err string) {
        components.Toast("Connection error", components.ToastError)
    },
    ReconnectInterval: 5 * time.Second,
    MaxReconnects:     10,
})

// Connect
wsStore.Connect()

// Subscribe to connection state
wsStore.Subscribe(func(state state.WSStoreState) {
    if state.Connected {
        connectionBadge.SetVariant(components.BadgeSuccess)
        connectionBadge.SetText("Online")
    } else if state.Connecting {
        connectionBadge.SetVariant(components.BadgeWarning)
        connectionBadge.SetText("Connecting...")
    } else {
        connectionBadge.SetVariant(components.BadgeError)
        connectionBadge.SetText("Offline")
    }
})

// Send typed messages
wsStore.SendTyped("chat.message", ChatMessage{Text: "Hello!"})

// Register handlers
wsStore.On("chat.message", func(data []byte) {
    var msg ChatMessage
    json.Unmarshal(data, &msg)
    appendMessage(msg)
})
```

## Best Practices

### 1. Always Handle Disconnections

```go
sub, err := posts.Subscribe(func(event api.PostEvent) {
    // Handle events
})
if err != nil {
    // Handle connection failure
    return
}

// Check connection periodically
go func() {
    for {
        time.Sleep(5 * time.Second)
        if !sub.IsConnected() {
            // Attempt reconnection or notify user
        }
    }
}()
```

### 2. Use Typed Events

```go
// Good: Type-safe event handling
ws.OnTyped(client, "post.created", func(post Post) {
    // Compile-time type checking
})

// Avoid: Manual JSON parsing
client.On("post.created", func(data json.RawMessage) {
    var post Post
    json.Unmarshal(data, &post) // Runtime errors
})
```

### 3. Clean Up Connections

```go
// Store subscription reference
var postsSub *api.Subscription

func startSubscription() {
    var err error
    postsSub, err = posts.Subscribe(handleEvent)
    // ...
}

func stopSubscription() {
    if postsSub != nil {
        postsSub.Close()
        postsSub = nil
    }
}
```

### 4. Debounce Rapid Updates

```go
var updateTimer *time.Timer

ws.OnTyped(client, "data.update", func(data DataUpdate) {
    if updateTimer != nil {
        updateTimer.Stop()
    }
    updateTimer = time.AfterFunc(100*time.Millisecond, func() {
        refreshUI(data)
    })
})
```

### 5. Handle Reconnection Gracefully

```go
wsStore := state.NewWebSocketStore(state.WebSocketConfig{
    URL:               "ws://localhost:8080/ws",
    ReconnectInterval: 5 * time.Second,
    MaxReconnects:     10,
    OnOpen: func() {
        // Resubscribe to channels after reconnect
        wsStore.SendTyped("subscribe", SubscribeRequest{
            Channels: []string{"posts", "users"},
        })
    },
})
```
