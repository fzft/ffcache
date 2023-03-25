package cache

import "time"

type Cacher interface {

	// Set sets the value for the given key. If the key already exists, it will be overwritten.
	Set([]byte, []byte, time.Duration) error

	// Get returns the value for the given key. If the key does not exist, it will return an error.
	Get([]byte) ([]byte, error)

	// Delete deletes the value for the given key. If the key does not exist, it will return an error.
	Delete([]byte) error

	// Has returns true if the key exists in the cache.
	Has([]byte) bool
}
