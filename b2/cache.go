package b2

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// DiskCache implements the Cache interface
type DiskCache struct {
	filename string
	mu       sync.RWMutex
}

// NewDiskCache returns a new disk based cache
//
// It is the caller's responsibility to make sure that
// the path exists and is writeable
func NewDiskCache(path string) (*DiskCache, error) {
	filename := filepath.Join(path, "cache.json")
	return &DiskCache{filename, sync.RWMutex{}}, nil
}

// Get returns the value from cache
func (c *DiskCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheBytes, err := ioutil.ReadFile(c.filename)
	if err != nil {
		// ignore "file does not exists" error
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(cacheBytes, &m)
	if err != nil {
		return nil, err
	}

	return m[key], nil
}

// Set stores value in cache
func (c *DiskCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := make(map[string]interface{})
	m[key] = value

	jsonBytes, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.filename, jsonBytes, 0600)
}

// InMemoryCache implements the Cache interface
type InMemoryCache struct {
	m  map[string]interface{}
	mu sync.RWMutex
}

// NewInMemoryCache returns a new in-memory cache
func NewInMemoryCache() (*InMemoryCache, error) {
	return &InMemoryCache{make(map[string]interface{}), sync.RWMutex{}}, nil
}

// Get returns the value from cache
func (c *InMemoryCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.m[key], nil
}

// Set stores value in cache
func (c *InMemoryCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return nil
}
