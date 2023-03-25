package cache

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

type Entry struct {
	Value []byte
	TTL   time.Time
}

type Cache struct {
	lock sync.RWMutex
	data map[string]*Entry
}

func New() *Cache {
	return &Cache{
		data: make(map[string]*Entry),
	}
}

// Set sets the value for a key with a given duration
func (c *Cache) Set(key []byte, val []byte, duration time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if duration == 0 {
		duration = math.MaxInt64
	}
	entry := &Entry{Value: val, TTL: time.Now().Add(duration)}
	c.data[string(key)] = entry

	log.Printf("Set key %s with value %s\n", string(key), string(val))
	return nil
}

// Get gets the value for a key, if the key is not found or expired, return error
func (c *Cache) Get(key []byte) ([]byte, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, ok := c.data[string(key)]
	if !ok {
		return nil, fmt.Errorf("key %s not found", string(key))
	}

	if time.Now().After(entry.TTL) {
		delete(c.data, string(key))
		return nil, fmt.Errorf("key %s expired", string(key))
	}
	log.Printf("Get key %s with value %s\n", string(key), string(entry.Value))
	return entry.Value, nil
}

func (c *Cache) Delete(key []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.data, string(key))
	log.Printf("Delete key %s\n", string(key))
	return nil
}

func (c *Cache) Has(bytes []byte) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	entry, ok := c.data[string(bytes)]

	if time.Now().After(entry.TTL) {
		delete(c.data, string(bytes))
		return false
	}
	log.Printf("Has key %s: %t\n", string(bytes), ok)
	return ok
}
