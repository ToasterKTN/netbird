package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/store"
)

// MemoryStoreType is the store type reported by the in-memory store. It keeps
// the value the go-cache backed store reported.
const MemoryStoreType = "go-cache"

const tagKeyPattern = "gocache_tag_%s"

var errMemoryNotFound = errors.New("value not found in memory store")

type memoryItem struct {
	value   any
	expires time.Time // zero means no expiry
}

func (i memoryItem) expired(now time.Time) bool {
	return !i.expires.IsZero() && now.After(i.expires)
}

// memoryStore is an in-memory TTL map. A TTL of 0 means the store's default
// timeout, a negative TTL means no expiry. Expired entries are hidden on read
// and swept at most once per cleanup interval on write.
type memoryStore struct {
	mu              sync.Mutex
	items           map[string]memoryItem
	tags            map[string]map[string]struct{}
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	lastSweep       time.Time
}

func newMemoryStore(maxTimeout, cleanupInterval time.Duration) Store {
	return &memoryStore{
		items:           make(map[string]memoryItem),
		tags:            make(map[string]map[string]struct{}),
		defaultTTL:      maxTimeout,
		cleanupInterval: cleanupInterval,
		lastSweep:       time.Now(),
	}
}

func (s *memoryStore) expiry(now time.Time, ttl time.Duration) time.Time {
	if ttl == 0 {
		ttl = s.defaultTTL
	}
	if ttl <= 0 {
		return time.Time{}
	}
	return now.Add(ttl)
}

// sweepLocked drops expired entries when the cleanup interval has passed.
func (s *memoryStore) sweepLocked(now time.Time) {
	if s.cleanupInterval <= 0 || now.Sub(s.lastSweep) < s.cleanupInterval {
		return
	}
	s.lastSweep = now
	for k, it := range s.items {
		if it.expired(now) {
			delete(s.items, k)
		}
	}
}

func (s *memoryStore) getLocked(key string, now time.Time) (memoryItem, bool) {
	it, ok := s.items[key]
	if !ok || it.expired(now) {
		return memoryItem{}, false
	}
	return it, true
}

func (s *memoryStore) Get(_ context.Context, key any) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.getLocked(key.(string), time.Now())
	if !ok {
		return nil, store.NotFoundWithCause(errMemoryNotFound)
	}
	return it.value, nil
}

func (s *memoryStore) GetWithTTL(_ context.Context, key any) (any, time.Duration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	it, ok := s.getLocked(key.(string), now)
	if !ok {
		return nil, 0, store.NotFoundWithCause(errMemoryNotFound)
	}
	if it.expires.IsZero() {
		return it.value, 0, nil
	}
	return it.value, it.expires.Sub(now), nil
}

func (s *memoryStore) Set(_ context.Context, key any, value any, options ...store.Option) error {
	opts := store.ApplyOptions(options...)
	k := key.(string)

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.sweepLocked(now)
	s.items[k] = memoryItem{value: value, expires: s.expiry(now, opts.Expiration)}
	for _, tag := range opts.Tags {
		tk := fmt.Sprintf(tagKeyPattern, tag)
		if s.tags[tk] == nil {
			s.tags[tk] = make(map[string]struct{})
		}
		s.tags[tk][k] = struct{}{}
	}
	return nil
}

func (s *memoryStore) Delete(_ context.Context, key any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key.(string))
	return nil
}

func (s *memoryStore) Invalidate(_ context.Context, options ...store.InvalidateOption) error {
	opts := store.ApplyInvalidateOptions(options...)

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, tag := range opts.Tags {
		tk := fmt.Sprintf(tagKeyPattern, tag)
		for k := range s.tags[tk] {
			delete(s.items, k)
		}
		delete(s.tags, tk)
	}
	return nil
}

func (s *memoryStore) Clear(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]memoryItem)
	s.tags = make(map[string]map[string]struct{})
	return nil
}

func (s *memoryStore) GetType() string {
	return MemoryStoreType
}

// SetNX stores value only when no unexpired entry exists under key.
func (s *memoryStore) SetNX(_ context.Context, key, value string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.sweepLocked(now)
	if _, ok := s.getLocked(key, now); ok {
		return false, nil
	}
	s.items[key] = memoryItem{value: value, expires: s.expiry(now, ttl)}
	return true, nil
}

// GetDel reads the value under key and removes it atomically, so no value is
// consumed twice.
func (s *memoryStore) GetDel(_ context.Context, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.getLocked(key, time.Now())
	if !ok {
		return "", false, nil
	}
	delete(s.items, key)

	str, isStr := it.value.(string)
	if !isStr {
		return "", false, fmt.Errorf("cached value is %T, not a string", it.value)
	}
	return str, true, nil
}
