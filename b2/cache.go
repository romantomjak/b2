package b2

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// DiskCache implements the Cache interface
type DiskCache struct {
	filename string
}

// NewDiskCache returns a new disk based cache
//
// It is the caller's responsibility to make sure that
// the path exists and is writeable
func NewDiskCache(path string) (*DiskCache, error) {
	filename := filepath.Join(path, "cache")
	return &DiskCache{filename}, nil
}

// Get returns the value from cache
func (c *DiskCache) Get(key string) (interface{}, error) {
	cacheBytes, err := ioutil.ReadFile(c.filename)
	if err != nil {
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
}

// NewInMemoryCache returns a new in-memory cache
func NewInMemoryCache() (*InMemoryCache, error) {
	return &InMemoryCache{}, nil
}

// Get returns the value from cache
func (c *InMemoryCache) Get(key string) (interface{}, error) {
	return nil, nil
}

// Set stores value in cache
func (c *InMemoryCache) Set(key string, value interface{}) error {
	return nil
}
