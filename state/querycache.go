//go:build js && wasm

package state

import (
	"sync"
	"time"
)

// QueryStatus represents the status of a query
type QueryStatus string

const (
	QueryIdle    QueryStatus = "idle"
	QueryLoading QueryStatus = "loading"
	QuerySuccess QueryStatus = "success"
	QueryError   QueryStatus = "error"
)

// QueryOptions configures a query
type QueryOptions struct {
	StaleTime      time.Duration // How long data stays fresh (default 0 = always stale)
	CacheTime      time.Duration // How long to keep data in cache (default 5 min)
	RefetchOnFocus bool          // Refetch when window regains focus
	RetryCount     int           // Number of retries on error (default 3)
	RetryDelay     time.Duration // Delay between retries (default 1s)
}

// QueryResult is returned from Query operations
type QueryResult struct {
	Data      any
	Error     error
	Status    QueryStatus
	IsLoading bool
	IsError   bool
	IsSuccess bool
	IsStale   bool
	Refetch   func()
}

// cacheEntry holds cached query data
type cacheEntry struct {
	data        any
	error       error
	status      QueryStatus
	lastFetched time.Time
	staleTime   time.Duration
	cacheTime   time.Duration
	subscribers []func()
}

// QueryCache manages cached queries (SWR-like behavior)
type QueryCache struct {
	mu             sync.RWMutex
	entries        map[string]*cacheEntry
	defaultOptions QueryOptions
}

var globalQueryCache *QueryCache

// GetQueryCache returns the global query cache
func GetQueryCache() *QueryCache {
	if globalQueryCache == nil {
		globalQueryCache = NewQueryCache(QueryOptions{
			StaleTime:  0,
			CacheTime:  5 * time.Minute,
			RetryCount: 3,
			RetryDelay: time.Second,
		})
	}
	return globalQueryCache
}

// NewQueryCache creates a new query cache
func NewQueryCache(defaultOptions QueryOptions) *QueryCache {
	return &QueryCache{
		entries:        make(map[string]*cacheEntry),
		defaultOptions: defaultOptions,
	}
}

// Query executes a query with caching
func (c *QueryCache) Query(key string, fetcher func() (any, error), options ...QueryOptions) *QueryResult {
	opts := c.defaultOptions
	if len(options) > 0 {
		opts = options[0]
	}

	c.mu.Lock()
	entry, exists := c.entries[key]

	if !exists {
		entry = &cacheEntry{
			status:    QueryIdle,
			staleTime: opts.StaleTime,
			cacheTime: opts.CacheTime,
		}
		c.entries[key] = entry
	}
	c.mu.Unlock()

	result := &QueryResult{
		Data:      entry.data,
		Error:     entry.error,
		Status:    entry.status,
		IsLoading: entry.status == QueryLoading,
		IsError:   entry.status == QueryError,
		IsSuccess: entry.status == QuerySuccess,
		IsStale:   c.isStale(entry),
	}

	// Create refetch function
	result.Refetch = func() {
		c.fetch(key, fetcher, opts)
	}

	// Fetch if no data or stale
	if !exists || c.isStale(entry) {
		go c.fetch(key, fetcher, opts)
	}

	return result
}

func (c *QueryCache) fetch(key string, fetcher func() (any, error), opts QueryOptions) {
	c.mu.Lock()
	entry := c.entries[key]
	entry.status = QueryLoading
	c.mu.Unlock()

	c.notifySubscribers(key)

	var lastErr error
	retries := opts.RetryCount
	if retries == 0 {
		retries = 1
	}

	for i := 0; i < retries; i++ {
		data, err := fetcher()
		if err == nil {
			c.mu.Lock()
			entry.data = data
			entry.error = nil
			entry.status = QuerySuccess
			entry.lastFetched = time.Now()
			c.mu.Unlock()
			c.notifySubscribers(key)
			return
		}
		lastErr = err
		if i < retries-1 {
			time.Sleep(opts.RetryDelay)
		}
	}

	c.mu.Lock()
	entry.error = lastErr
	entry.status = QueryError
	c.mu.Unlock()
	c.notifySubscribers(key)
}

func (c *QueryCache) isStale(entry *cacheEntry) bool {
	if entry.status != QuerySuccess {
		return true
	}
	if entry.staleTime == 0 {
		return true // Always stale if staleTime is 0
	}
	return time.Since(entry.lastFetched) > entry.staleTime
}

func (c *QueryCache) notifySubscribers(key string) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	if !exists {
		c.mu.RUnlock()
		return
	}
	subscribers := make([]func(), len(entry.subscribers))
	copy(subscribers, entry.subscribers)
	c.mu.RUnlock()

	for _, fn := range subscribers {
		fn()
	}
}

// Subscribe adds a subscriber to a query
func (c *QueryCache) Subscribe(key string, fn func()) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		entry = &cacheEntry{status: QueryIdle}
		c.entries[key] = entry
	}

	entry.subscribers = append(entry.subscribers, fn)

	// Return unsubscribe function
	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		for i, subscriber := range entry.subscribers {
			// Compare function pointers
			if &subscriber == &fn {
				entry.subscribers = append(entry.subscribers[:i], entry.subscribers[i+1:]...)
				break
			}
		}
	}
}

// Invalidate invalidates a cached query, causing refetch on next access
func (c *QueryCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[key]; exists {
		entry.lastFetched = time.Time{} // Reset fetch time to force refetch
	}
}

// InvalidateAll invalidates all cached queries
func (c *QueryCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, entry := range c.entries {
		entry.lastFetched = time.Time{}
	}
}

// SetData manually sets data in the cache (for optimistic updates)
func (c *QueryCache) SetData(key string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		entry = &cacheEntry{status: QueryIdle}
		c.entries[key] = entry
	}

	entry.data = data
	entry.status = QuerySuccess
	entry.lastFetched = time.Now()
}

// GetData returns cached data without fetching
func (c *QueryCache) GetData(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || entry.status != QuerySuccess {
		return nil, false
	}
	return entry.data, true
}

// Clear removes a specific entry from cache
func (c *QueryCache) Clear(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// ClearAll removes all entries from cache
func (c *QueryCache) ClearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
}

// Prefetch fetches data in advance
func (c *QueryCache) Prefetch(key string, fetcher func() (any, error), options ...QueryOptions) {
	opts := c.defaultOptions
	if len(options) > 0 {
		opts = options[0]
	}

	c.mu.Lock()
	_, exists := c.entries[key]
	if !exists {
		c.entries[key] = &cacheEntry{
			status:    QueryIdle,
			staleTime: opts.StaleTime,
			cacheTime: opts.CacheTime,
		}
	}
	c.mu.Unlock()

	go c.fetch(key, fetcher, opts)
}

// UseQuery is a convenience function for querying with the global cache
func UseQuery(key string, fetcher func() (any, error), options ...QueryOptions) *QueryResult {
	return GetQueryCache().Query(key, fetcher, options...)
}
