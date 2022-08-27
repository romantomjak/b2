package b2

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// DiskCache implements the Cache interface.
type DiskCache struct {
	filename string
	mu       sync.RWMutex
}

// NewDiskCache returns a new disk based cache.
//
// It is the caller's responsibility to make sure that the path exists
// and is writeable.
func NewDiskCache(path string) (*DiskCache, error) {
	filename := filepath.Join(path, "cache.json")
	return &DiskCache{filename, sync.RWMutex{}}, nil
}

func (c *DiskCache) Get(key string, value interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cacheBytes, err := ioutil.ReadFile(c.filename)
	if err != nil {
		// ignore "file does not exists" error
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(cacheBytes, &m); err != nil {
		return err
	}

	// Admittedly, marshaling and unmarshaling the cache value to json is
	// fairly strange, but that's because we first read the whole cache file
	// from disk in order to get the data pointed to by key. Then to avoid
	// reimplementing what json.Unmarshal does when decoding values, we just
	// marshal the whole data to json and then unmarshal it again into the
	// value pointed to by value.
	bytes, err := json.Marshal(m[key])
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

func (c *DiskCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := make(map[string]interface{})
	m[key] = value

	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.filename, jsonBytes, 0600)
}

// InMemoryCache implements the Cache interface.
type InMemoryCache struct {
	m  map[string]interface{}
	mu sync.RWMutex
}

// NewInMemoryCache returns a new in-memory cache.
func NewInMemoryCache() (*InMemoryCache, error) {
	return &InMemoryCache{make(map[string]interface{}), sync.RWMutex{}}, nil
}

func (c *InMemoryCache) Get(key string, value interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Admittedly, marshaling and unmarshaling the cache value to json is
	// fairly strange, but that's because we first read the whole cache file
	// from disk in order to get the data pointed to by key. Then to avoid
	// reimplementing what json.Unmarshal does when decoding values, we just
	// marshal the whole data to json and then unmarshal it again into the
	// value pointed to by value.
	bytes, err := json.Marshal(c.m[key])
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

func (c *InMemoryCache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
	return nil
}
