# State Management

Gux provides a comprehensive state management system with reactive stores, browser storage integration, async data handling, and query caching.

## Store

The core `Store[T]` type provides generic reactive state management.

### Creating a Store

```go
import "github.com/dougbarrett/guxstate"

// Create with initial value
type AppState struct {
    User     *User
    Count    int
    Theme    string
    Settings map[string]bool
}

store := state.New(AppState{
    User:     nil,
    Count:    0,
    Theme:    "light",
    Settings: make(map[string]bool),
})
```

### Reading State

```go
// Get current state (read-only snapshot)
current := store.Get()
fmt.Println(current.Count) // 0
fmt.Println(current.Theme) // "light"
```

### Updating State

```go
// Replace entire state
store.Set(AppState{
    Count: 10,
    Theme: "dark",
})

// Mutate state (recommended)
store.Update(func(s *AppState) {
    s.Count++
})

store.Update(func(s *AppState) {
    s.Theme = "dark"
    s.Settings["notifications"] = true
})
```

### Subscribing to Changes

```go
// Subscribe returns an unsubscribe function
unsubscribe := store.Subscribe(func(s AppState) {
    fmt.Println("State changed:", s.Count)
    // Update UI, trigger side effects, etc.
})

// Clean up when done
defer unsubscribe()

// Multiple subscribers are supported
store.Subscribe(func(s AppState) {
    // Another subscriber
})
```

### Derived Stores

Create computed stores that automatically update:

```go
// Parent store
userStore := state.New(User{Name: "John", Age: 30})

// Derived store that only contains the name
nameStore := state.Derived(userStore, func(u User) string {
    return u.Name
})

// Subscribe to derived store
nameStore.Subscribe(func(name string) {
    fmt.Println("Name changed:", name)
})

// Updates to parent propagate to derived
userStore.Update(func(u *User) {
    u.Name = "Jane" // Triggers nameStore subscribers
})
```

## Browser Storage

### Raw Storage Access

```go
// LocalStorage (persists across sessions)
local := state.LocalStorage()
local.Set("key", "value")
value := local.Get("key")
local.Remove("key")
local.Clear()

// Check existence
if local.Has("key") { ... }

// Get all keys
keys := local.Keys()

// SessionStorage (cleared when browser closes)
session := state.SessionStorage()
// Same API as LocalStorage
```

### JSON Storage

```go
// Store complex objects
user := User{Name: "John", Email: "john@example.com"}
state.SetLocalJSON("currentUser", user)

// Retrieve
var loaded User
err := state.GetLocalJSON("currentUser", &loaded)

// Session storage equivalents
state.SetSessionJSON("tempData", data)
state.GetSessionJSON("tempData", &result)
```

### Convenience Functions

```go
// String values
state.SetLocalString("theme", "dark")
theme := state.GetLocalString("theme", "light") // with default

state.SetSessionString("token", authToken)
token := state.GetSessionString("token", "")

// Remove
state.RemoveLocal("key")
state.RemoveSession("key")
```

## Persistent Store

Automatically saves to localStorage on every change:

```go
// Create persistent store - loads from localStorage if exists
userStore := state.NewPersistentStore("currentUser", User{
    Name:  "Guest",
    Theme: "light",
})

// Updates automatically save to localStorage
userStore.Update(func(u *User) {
    u.Name = "John"
    u.Theme = "dark"
})
// localStorage["currentUser"] = {"Name":"John","Theme":"dark"}

// Subscribe works normally
userStore.Subscribe(func(u User) {
    // React to changes
})

// On page reload, state is restored from localStorage
```

## Session Store

Like PersistentStore, but uses sessionStorage (clears on browser close):

```go
// Good for temporary session data
cartStore := state.NewSessionStore("shoppingCart", Cart{
    Items: []CartItem{},
})

cartStore.Update(func(c *Cart) {
    c.Items = append(c.Items, CartItem{ID: 1, Qty: 2})
})
// sessionStorage["shoppingCart"] = {...}

// Cleared automatically when browser closes
```

## AsyncStore

Manages async data loading with loading/error states:

```go
// Create async store
postsStore := state.NewAsync[[]Post]()

// With default data
postsStore := state.NewAsyncWithDefault([]Post{})

// Load data
postsStore.Load(func() ([]Post, error) {
    return api.GetPosts() // Your async operation
})

// Check states
if postsStore.IsLoading() {
    // Show spinner
}

if postsStore.HasError() {
    err := postsStore.Err()
    // Show error message
}

// Get data
posts := postsStore.Data()

// Set data directly (bypass loading)
postsStore.SetData([]Post{...})

// Set error directly
postsStore.SetError(errors.New("failed"))
```

### Using with UI

```go
func renderPosts(store *state.AsyncStore[[]Post]) js.Value {
    container := components.Div("space-y-4")

    if store.IsLoading() {
        return components.Spinner(components.SpinnerProps{
            Label: "Loading posts...",
        })
    }

    if store.HasError() {
        return components.Alert(components.AlertProps{
            Variant: components.AlertError,
            Message: store.Err().Error(),
        })
    }

    for _, post := range store.Data() {
        card := components.Card(components.CardProps{},
            components.H3(post.Title),
            components.Text(post.Body),
        )
        container.Call("appendChild", card)
    }

    return container
}
```

### Subscribing to AsyncStore

```go
postsStore.Subscribe(func(state state.AsyncState[[]Post]) {
    if state.Loading {
        showLoading()
    } else if state.Error != nil {
        showError(state.Error)
    } else {
        renderPosts(state.Data)
    }
})
```

## Query Cache

SWR-style (Stale-While-Revalidate) caching for data fetching:

### Basic Usage

```go
// Get global cache instance
cache := state.GetQueryCache()

// Or use convenience function
result := state.UseQuery("posts", func() (any, error) {
    return api.GetPosts()
}, state.QueryOptions{
    StaleTime: 5 * time.Minute,
})

// Access result
if result.IsLoading {
    // Show spinner
}
if result.IsError {
    // Handle error
}
if result.IsSuccess {
    posts := result.Data.([]Post)
}

// Manual refetch
result.Refetch()
```

### Query Options

```go
state.QueryOptions{
    // How long until data is considered stale
    StaleTime: 5 * time.Minute,

    // How long to keep data in cache (default 5 min)
    CacheTime: 10 * time.Minute,

    // Refetch when window gains focus
    RefetchOnFocus: true,

    // Number of retry attempts (default 3)
    RetryCount: 3,

    // Delay between retries (default 1s)
    RetryDelay: time.Second,
}
```

### Query Result

```go
type QueryResult struct {
    Data      any          // Cached data
    Error     error        // Error if any
    Status    QueryStatus  // idle, loading, success, error
    IsLoading bool
    IsError   bool
    IsSuccess bool
    IsStale   bool         // Data exists but is stale
    Refetch   func()       // Manually refetch
}
```

### Cache Operations

```go
cache := state.GetQueryCache()

// Invalidate specific query (marks as stale)
cache.Invalidate("posts")

// Invalidate all queries
cache.InvalidateAll()

// Set data directly (optimistic updates)
cache.SetData("posts", updatedPosts)

// Get cached data
data, exists := cache.GetData("posts")

// Clear specific query
cache.Clear("posts")

// Clear all cache
cache.ClearAll()

// Prefetch data
cache.Prefetch("posts", fetchPosts, options)
```

### Subscribing to Cache Changes

```go
// Watch specific query
unsubscribe := cache.Subscribe("posts", func(result *state.QueryResult) {
    if result.IsSuccess {
        updateUI(result.Data)
    }
})
defer unsubscribe()
```

### Optimistic Updates

```go
// Update cache immediately, then sync with server
cache.SetData("posts", append(currentPosts, newPost))

// Make API call
_, err := api.CreatePost(newPost)
if err != nil {
    // Revert on error
    cache.SetData("posts", currentPosts)
    cache.Invalidate("posts") // Refetch from server
}
```

## WebSocket Store

Integrated WebSocket with state management:

```go
wsStore := state.NewWebSocketStore(state.WebSocketConfig{
    URL: "ws://localhost:8080/ws",
    OnOpen: func() {
        fmt.Println("Connected")
    },
    OnClose: func(code int, reason string) {
        fmt.Println("Disconnected")
    },
    OnMessage: func(data []byte) {
        // Handle raw messages
    },
    OnError: func(err string) {
        fmt.Println("Error:", err)
    },
    ReconnectInterval: 5 * time.Second,
    MaxReconnects:     10,
})

// Connect
wsStore.Connect()

// Send messages
wsStore.Send([]byte("hello"))
wsStore.SendJSON(MyMessage{Type: "ping"})
wsStore.SendTyped("chat.message", ChatMessage{Text: "Hi"})

// Register typed handlers
wsStore.On("chat.message", func(data []byte) {
    var msg ChatMessage
    json.Unmarshal(data, &msg)
    // Handle message
})

// Access state
state := wsStore.State()
fmt.Println(state.Connected)    // bool
fmt.Println(state.Connecting)   // bool
fmt.Println(state.LastMessage)  // string
fmt.Println(state.MessageCount) // int
fmt.Println(state.Error)        // string

// Subscribe to state changes
wsStore.Subscribe(func(s state.WSStoreState) {
    if s.Connected {
        updateConnectionStatus("online")
    } else {
        updateConnectionStatus("offline")
    }
})

// Get underlying store for advanced use
store := wsStore.Store()

// Message history
messages := wsStore.Messages()
wsStore.ClearMessages()

// Check connection
if wsStore.IsConnected() {
    // Send message
}

// Close
wsStore.Close()
```

## Best Practices

### 1. Single Source of Truth

```go
// Create a central app store
var appStore = state.New(AppState{})

// Access from anywhere
func getCurrentUser() *User {
    return appStore.Get().User
}
```

### 2. Keep State Normalized

```go
// Instead of nested data
type BadState struct {
    Users []User // Each user has []Post
}

// Keep collections flat
type GoodState struct {
    Users map[int]User
    Posts map[int]Post
}
```

### 3. Use Derived Stores for Computed Values

```go
// Don't recompute in subscribers
userStore.Subscribe(func(users []User) {
    // Bad: computing every time
    activeCount := 0
    for _, u := range users {
        if u.Active { activeCount++ }
    }
})

// Use derived store instead
activeCountStore := state.Derived(userStore, func(users []User) int {
    count := 0
    for _, u := range users {
        if u.Active { count++ }
    }
    return count
})
```

### 4. Clean Up Subscriptions

```go
// Always store unsubscribe functions
unsub1 := store1.Subscribe(...)
unsub2 := store2.Subscribe(...)

// Clean up when component unmounts
cleanup := func() {
    unsub1()
    unsub2()
}
```

### 5. Use AsyncStore for API Data

```go
// Don't manage loading state manually
var loading bool
var error error
var data []Post

func loadPosts() {
    loading = true
    data, err = api.GetPosts()
    loading = false
    error = err
}

// Use AsyncStore instead
postsStore := state.NewAsync[[]Post]()
postsStore.Load(api.GetPosts)
```

### 6. Leverage Query Cache for Repeated Fetches

```go
// Cache prevents redundant API calls
result := state.UseQuery("user-"+id, func() (any, error) {
    return api.GetUser(id)
}, state.QueryOptions{
    StaleTime: 5 * time.Minute,
})
```
