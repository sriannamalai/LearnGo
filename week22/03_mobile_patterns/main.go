package main

// ========================================
// Week 22 — Lesson 3: Mobile Patterns
// ========================================
// This lesson covers essential mobile development patterns:
//   - Persistent storage (preferences and file-based)
//   - Network-aware operations
//   - Offline-first patterns
//   - Push notification concepts
//   - Background tasks and lifecycle management
//
// These patterns apply whether you're using Fyne mobile,
// gomobile bind, or any Go-based mobile approach.
//
// Run:
//   go run .

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ========================================
// Pattern 1: Persistent Storage
// ========================================
// Mobile apps need to persist data between launches.
// Two common approaches:
//   1. Key-value preferences (settings, tokens, flags)
//   2. File-based storage (JSON, SQLite, etc.)
//
// In Fyne, use app.Preferences() for simple values:
//   app.Preferences().SetString("username", "alice")
//   name := app.Preferences().String("username")
//   app.Preferences().SetBool("darkMode", true)
//   dark := app.Preferences().Bool("darkMode")

// PreferencesStore provides a file-backed key-value store.
// This simulates what Fyne's Preferences API does internally.
type PreferencesStore struct {
	mu       sync.RWMutex
	data     map[string]interface{}
	filePath string
}

// NewPreferencesStore creates a new preferences store backed by a JSON file.
func NewPreferencesStore(filePath string) *PreferencesStore {
	ps := &PreferencesStore{
		data:     make(map[string]interface{}),
		filePath: filePath,
	}
	ps.load()
	return ps
}

func (ps *PreferencesStore) load() {
	data, err := os.ReadFile(ps.filePath)
	if err != nil {
		return // File doesn't exist yet
	}
	json.Unmarshal(data, &ps.data)
}

// Save persists the preferences to disk.
func (ps *PreferencesStore) Save() error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	dir := filepath.Dir(ps.filePath)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(ps.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ps.filePath, data, 0644)
}

// SetString stores a string value.
func (ps *PreferencesStore) SetString(key, value string) {
	ps.mu.Lock()
	ps.data[key] = value
	ps.mu.Unlock()
}

// GetString retrieves a string value with a default.
func (ps *PreferencesStore) GetString(key, defaultValue string) string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if val, ok := ps.data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// SetBool stores a boolean value.
func (ps *PreferencesStore) SetBool(key string, value bool) {
	ps.mu.Lock()
	ps.data[key] = value
	ps.mu.Unlock()
}

// GetBool retrieves a boolean value with a default.
func (ps *PreferencesStore) GetBool(key string, defaultValue bool) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if val, ok := ps.data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// ========================================
// Pattern 2: Offline-First Data Store
// ========================================
// Offline-first means the app works without a network
// connection and syncs when connectivity returns.
//
// Strategy:
//   1. Always read from local cache first
//   2. Display cached data immediately
//   3. Fetch fresh data in the background
//   4. Update the cache when new data arrives
//   5. Handle conflicts (last-write-wins or merge)

// CacheEntry represents a cached item with metadata.
type CacheEntry struct {
	Key       string      `json:"key"`
	Data      interface{} `json:"data"`
	CachedAt  time.Time   `json:"cachedAt"`
	ExpiresAt time.Time   `json:"expiresAt"`
	Version   int         `json:"version"`
}

// OfflineCache provides an offline-first caching layer.
type OfflineCache struct {
	mu       sync.RWMutex
	entries  map[string]CacheEntry
	filePath string
	ttl      time.Duration
}

// NewOfflineCache creates a new offline cache.
func NewOfflineCache(filePath string, ttl time.Duration) *OfflineCache {
	cache := &OfflineCache{
		entries:  make(map[string]CacheEntry),
		filePath: filePath,
		ttl:      ttl,
	}
	cache.loadFromDisk()
	return cache
}

func (oc *OfflineCache) loadFromDisk() {
	data, err := os.ReadFile(oc.filePath)
	if err != nil {
		return
	}

	var entries []CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return
	}

	for _, entry := range entries {
		oc.entries[entry.Key] = entry
	}
}

// SaveToDisk persists the cache to disk.
func (oc *OfflineCache) SaveToDisk() error {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	dir := filepath.Dir(oc.filePath)
	os.MkdirAll(dir, 0755)

	var entries []CacheEntry
	for _, entry := range oc.entries {
		entries = append(entries, entry)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(oc.filePath, data, 0644)
}

// Set stores data in the cache.
func (oc *OfflineCache) Set(key string, data interface{}) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	version := 1
	if existing, ok := oc.entries[key]; ok {
		version = existing.Version + 1
	}

	oc.entries[key] = CacheEntry{
		Key:       key,
		Data:      data,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(oc.ttl),
		Version:   version,
	}
}

// Get retrieves data from the cache.
// Returns the data, whether it was found, and whether it has expired.
func (oc *OfflineCache) Get(key string) (interface{}, bool, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()

	entry, found := oc.entries[key]
	if !found {
		return nil, false, false
	}

	expired := time.Now().After(entry.ExpiresAt)
	return entry.Data, true, expired
}

// GetOrFetch retrieves cached data or fetches fresh data.
// This is the core offline-first pattern:
//   1. Check cache
//   2. If found and fresh, return cached
//   3. If found but stale, return cached AND fetch in background
//   4. If not found, fetch synchronously
func (oc *OfflineCache) GetOrFetch(key string, fetcher func() (interface{}, error)) (interface{}, error) {
	data, found, expired := oc.Get(key)

	if found && !expired {
		// Cache hit — fresh data
		fmt.Printf("  [Cache] HIT (fresh): %s\n", key)
		return data, nil
	}

	if found && expired {
		// Cache hit — stale data, refresh in background
		fmt.Printf("  [Cache] HIT (stale): %s — refreshing...\n", key)
		go func() {
			freshData, err := fetcher()
			if err == nil {
				oc.Set(key, freshData)
				oc.SaveToDisk()
				fmt.Printf("  [Cache] REFRESHED: %s\n", key)
			}
		}()
		return data, nil // Return stale data immediately
	}

	// Cache miss — fetch synchronously
	fmt.Printf("  [Cache] MISS: %s — fetching...\n", key)
	freshData, err := fetcher()
	if err != nil {
		return nil, err
	}

	oc.Set(key, freshData)
	oc.SaveToDisk()
	return freshData, nil
}

// ========================================
// Pattern 3: Network-Aware Operations
// ========================================
// Mobile apps must handle varying network conditions:
//   - WiFi, cellular, offline
//   - Slow or intermittent connections
//   - Data usage concerns on cellular

// NetworkStatus represents the current connectivity state.
type NetworkStatus int

const (
	NetworkOffline  NetworkStatus = iota
	NetworkCellular               // Limited data
	NetworkWiFi                   // Unlimited data
)

func (ns NetworkStatus) String() string {
	switch ns {
	case NetworkOffline:
		return "Offline"
	case NetworkCellular:
		return "Cellular"
	case NetworkWiFi:
		return "WiFi"
	default:
		return "Unknown"
	}
}

// NetworkAwareClient adapts behavior based on connectivity.
type NetworkAwareClient struct {
	status     NetworkStatus
	cache      *OfflineCache
	mu         sync.RWMutex
	listeners  []func(NetworkStatus)
}

// NewNetworkAwareClient creates a new network-aware client.
func NewNetworkAwareClient(cache *OfflineCache) *NetworkAwareClient {
	return &NetworkAwareClient{
		status: NetworkWiFi, // Default assumption
		cache:  cache,
	}
}

// SetStatus updates the network status and notifies listeners.
// On a real device, this would be called by the OS connectivity
// framework (Reachability on iOS, ConnectivityManager on Android).
func (nac *NetworkAwareClient) SetStatus(status NetworkStatus) {
	nac.mu.Lock()
	oldStatus := nac.status
	nac.status = status
	nac.mu.Unlock()

	if oldStatus != status {
		fmt.Printf("  [Network] Status changed: %s -> %s\n",
			oldStatus, status)
		for _, listener := range nac.listeners {
			listener(status)
		}
	}
}

// OnStatusChange registers a callback for network changes.
func (nac *NetworkAwareClient) OnStatusChange(fn func(NetworkStatus)) {
	nac.listeners = append(nac.listeners, fn)
}

// GetStatus returns the current network status.
func (nac *NetworkAwareClient) GetStatus() NetworkStatus {
	nac.mu.RLock()
	defer nac.mu.RUnlock()
	return nac.status
}

// FetchData demonstrates network-aware data fetching.
func (nac *NetworkAwareClient) FetchData(key string) (interface{}, error) {
	status := nac.GetStatus()

	switch status {
	case NetworkOffline:
		// Offline: use cache only
		data, found, _ := nac.cache.Get(key)
		if !found {
			return nil, fmt.Errorf("offline and no cached data for %q", key)
		}
		fmt.Printf("  [Network] Offline — serving cached: %s\n", key)
		return data, nil

	case NetworkCellular:
		// Cellular: use cache if available, fetch only if needed
		data, found, expired := nac.cache.Get(key)
		if found && !expired {
			fmt.Printf("  [Network] Cellular — serving cached: %s\n", key)
			return data, nil
		}
		// On cellular, we might skip large downloads
		if found {
			fmt.Printf("  [Network] Cellular — serving stale cache: %s\n", key)
			return data, nil
		}
		fmt.Printf("  [Network] Cellular — fetching (minimal): %s\n", key)
		return nac.fetchFromServer(key)

	case NetworkWiFi:
		// WiFi: always try fresh data
		return nac.cache.GetOrFetch(key, func() (interface{}, error) {
			return nac.fetchFromServer(key)
		})
	}

	return nil, fmt.Errorf("unknown network status")
}

func (nac *NetworkAwareClient) fetchFromServer(key string) (interface{}, error) {
	// Simulate network request
	fmt.Printf("  [Network] Fetching from server: %s\n", key)
	return map[string]string{
		"key":       key,
		"fetchedAt": time.Now().Format(time.RFC3339),
		"source":    "server",
	}, nil
}

// ========================================
// Pattern 4: Sync Queue
// ========================================
// When offline, queue operations to sync later.
// This ensures no data is lost.

// SyncOperation represents a queued operation.
type SyncOperation struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"` // "create", "update", "delete"
	Resource  string      `json:"resource"`
	Data      interface{} `json:"data"`
	QueuedAt  time.Time   `json:"queuedAt"`
	Attempts  int         `json:"attempts"`
	LastError string      `json:"lastError,omitempty"`
}

// SyncQueue manages pending operations for offline sync.
type SyncQueue struct {
	mu         sync.Mutex
	operations []SyncOperation
	filePath   string
	nextID     int
}

// NewSyncQueue creates a new sync queue.
func NewSyncQueue(filePath string) *SyncQueue {
	sq := &SyncQueue{
		filePath: filePath,
		nextID:   1,
	}
	sq.loadFromDisk()
	return sq
}

func (sq *SyncQueue) loadFromDisk() {
	data, err := os.ReadFile(sq.filePath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &sq.operations)
	if len(sq.operations) > 0 {
		sq.nextID = len(sq.operations) + 1
	}
}

// Enqueue adds an operation to the sync queue.
func (sq *SyncQueue) Enqueue(opType, resource string, data interface{}) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	op := SyncOperation{
		ID:       fmt.Sprintf("op_%d", sq.nextID),
		Type:     opType,
		Resource: resource,
		Data:     data,
		QueuedAt: time.Now(),
	}
	sq.operations = append(sq.operations, op)
	sq.nextID++
	sq.saveToDisk()

	fmt.Printf("  [SyncQueue] Enqueued: %s %s\n", opType, resource)
}

// ProcessQueue attempts to sync all pending operations.
func (sq *SyncQueue) ProcessQueue(processor func(SyncOperation) error) int {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if len(sq.operations) == 0 {
		return 0
	}

	fmt.Printf("  [SyncQueue] Processing %d operations...\n", len(sq.operations))
	synced := 0
	var remaining []SyncOperation

	for _, op := range sq.operations {
		err := processor(op)
		if err != nil {
			op.Attempts++
			op.LastError = err.Error()
			remaining = append(remaining, op)
			fmt.Printf("  [SyncQueue] Failed: %s (%s)\n", op.ID, err)
		} else {
			synced++
			fmt.Printf("  [SyncQueue] Synced: %s %s\n", op.Type, op.Resource)
		}
	}

	sq.operations = remaining
	sq.saveToDisk()
	return synced
}

// PendingCount returns the number of unsynced operations.
func (sq *SyncQueue) PendingCount() int {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return len(sq.operations)
}

func (sq *SyncQueue) saveToDisk() {
	dir := filepath.Dir(sq.filePath)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(sq.operations, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(sq.filePath, data, 0644)
}

// ========================================
// Pattern 5: Push Notification Concepts
// ========================================
// Push notifications require platform-specific setup:
//
// iOS (APNs — Apple Push Notification service):
//   1. Register for push in AppDelegate
//   2. Receive device token
//   3. Send token to your Go server
//   4. Server uses APNs API to send notifications
//
// Android (FCM — Firebase Cloud Messaging):
//   1. Add Firebase SDK to Android project
//   2. Implement FirebaseMessagingService
//   3. Receive FCM registration token
//   4. Send token to your Go server
//   5. Server uses FCM API to send notifications
//
// Go Server-Side (both platforms):
//
//   type PushNotification struct {
//       Title    string `json:"title"`
//       Body     string `json:"body"`
//       Data     map[string]string `json:"data,omitempty"`
//       Badge    int    `json:"badge,omitempty"`
//       Sound    string `json:"sound,omitempty"`
//   }
//
//   // Store device tokens
//   type DeviceToken struct {
//       UserID   string
//       Platform string // "ios" or "android"
//       Token    string
//   }
//
//   // Send via APNs (iOS)
//   func sendAPNs(token string, notification PushNotification) error {
//       // Use a library like github.com/sideshow/apns2
//       return nil
//   }
//
//   // Send via FCM (Android)
//   func sendFCM(token string, notification PushNotification) error {
//       // Use Firebase Admin SDK or HTTP API
//       return nil
//   }

// ========================================
// Main — Demonstrate All Patterns
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 22 - Lesson 3: Mobile Patterns")
	fmt.Println("========================================")
	fmt.Println()

	// Use temp directory for demo files
	tmpDir := filepath.Join(os.TempDir(), "learngo-mobile-patterns")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// ========================================
	// Demo 1: Preferences
	// ========================================
	fmt.Println("--- Pattern 1: Persistent Storage ---")
	prefs := NewPreferencesStore(filepath.Join(tmpDir, "prefs.json"))

	prefs.SetString("username", "gopher")
	prefs.SetString("theme", "dark")
	prefs.SetBool("notifications", true)
	prefs.SetBool("biometricAuth", false)
	prefs.Save()

	fmt.Printf("  Username: %s\n", prefs.GetString("username", ""))
	fmt.Printf("  Theme: %s\n", prefs.GetString("theme", "light"))
	fmt.Printf("  Notifications: %v\n", prefs.GetBool("notifications", false))
	fmt.Printf("  Missing key: %q\n", prefs.GetString("missing", "default_value"))

	// ========================================
	// Demo 2: Offline Cache
	// ========================================
	fmt.Println("\n--- Pattern 2: Offline-First Cache ---")
	cache := NewOfflineCache(
		filepath.Join(tmpDir, "cache.json"),
		5*time.Minute, // Cache entries expire after 5 minutes
	)

	// First access — cache miss, fetches data
	cache.GetOrFetch("user_profile", func() (interface{}, error) {
		return map[string]string{"name": "Alice", "email": "alice@example.com"}, nil
	})

	// Second access — cache hit
	cache.GetOrFetch("user_profile", func() (interface{}, error) {
		return map[string]string{"name": "Alice Updated"}, nil
	})

	// Different key — cache miss
	cache.GetOrFetch("settings", func() (interface{}, error) {
		return map[string]string{"language": "en", "timezone": "UTC"}, nil
	})

	// ========================================
	// Demo 3: Network-Aware Client
	// ========================================
	fmt.Println("\n--- Pattern 3: Network-Aware Operations ---")
	client := NewNetworkAwareClient(cache)

	// Register status change listener
	client.OnStatusChange(func(status NetworkStatus) {
		fmt.Printf("  [Listener] Network is now: %s\n", status)
	})

	fmt.Println("\nOn WiFi:")
	client.SetStatus(NetworkWiFi)
	client.FetchData("user_profile")

	fmt.Println("\nOn Cellular:")
	client.SetStatus(NetworkCellular)
	client.FetchData("user_profile") // Uses cache to save data

	fmt.Println("\nOffline:")
	client.SetStatus(NetworkOffline)
	client.FetchData("user_profile") // Uses cache only

	fmt.Println("\nOffline (uncached key):")
	_, err := client.FetchData("uncached_key")
	if err != nil {
		fmt.Printf("  Expected error: %v\n", err)
	}

	// ========================================
	// Demo 4: Sync Queue
	// ========================================
	fmt.Println("\n--- Pattern 4: Sync Queue ---")
	syncQueue := NewSyncQueue(filepath.Join(tmpDir, "sync_queue.json"))

	// Simulate offline operations
	fmt.Println("Queuing operations while offline...")
	syncQueue.Enqueue("create", "/api/notes", map[string]string{
		"title": "New Note", "content": "Written offline",
	})
	syncQueue.Enqueue("update", "/api/notes/1", map[string]string{
		"title": "Updated Note",
	})
	syncQueue.Enqueue("delete", "/api/notes/5", nil)

	fmt.Printf("  Pending operations: %d\n", syncQueue.PendingCount())

	// Simulate coming back online
	fmt.Println("\nBack online — syncing...")
	synced := syncQueue.ProcessQueue(func(op SyncOperation) error {
		// Simulate successful sync for most operations
		if op.Resource == "/api/notes/5" {
			return fmt.Errorf("server error: 500")
		}
		return nil
	})

	fmt.Printf("  Synced: %d, Remaining: %d\n", synced, syncQueue.PendingCount())

	// Retry failed operations
	fmt.Println("\nRetrying failed operations...")
	synced = syncQueue.ProcessQueue(func(op SyncOperation) error {
		return nil // Succeeds this time
	})
	fmt.Printf("  Synced: %d, Remaining: %d\n", synced, syncQueue.PendingCount())

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n--- Summary: Mobile Patterns ---")
	fmt.Println("  1. Preferences: Key-value storage for settings")
	fmt.Println("  2. Offline Cache: Always show data, refresh in background")
	fmt.Println("  3. Network Aware: Adapt behavior to connectivity")
	fmt.Println("  4. Sync Queue: Queue offline changes, sync when online")
	fmt.Println("  5. Push Notifications: Server-side with APNs/FCM")
}
