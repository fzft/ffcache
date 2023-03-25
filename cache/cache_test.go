package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCache_Set(t *testing.T) {
	c := New()
	err := c.Set([]byte("key"), []byte("value"), 0)
	if err != nil {
		t.Errorf("Error setting key: %s", err)
	}
}

func TestCache_Get(t *testing.T) {
	c := New()
	err := c.Set([]byte("key"), []byte("value"), 0)
	assert.Nil(t, err)

	value, err := c.Get([]byte("key"))
	assert.Nil(t, err)

	assert.Equal(t, []byte("value"), value)
}

func TestCache_Delete(t *testing.T) {
	c := New()
	err := c.Set([]byte("key"), []byte("value"), 0)
	if err != nil {
		t.Errorf("Error setting key: %s", err)
	}

	err = c.Delete([]byte("key"))
	if err != nil {
		t.Errorf("Error deleting key: %s", err)
	}

	_, err = c.Get([]byte("key"))
	if err == nil {
		t.Errorf("Expected error getting key, got nil")
	}
}

func TestCache_Has(t *testing.T) {
	c := New()
	err := c.Set([]byte("key"), []byte("value"), 0)
	if err != nil {
		t.Errorf("Error setting key: %s", err)
	}

	if !c.Has([]byte("key")) {
		t.Errorf("Expected key to exist")
	}
}
