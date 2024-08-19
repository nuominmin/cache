package cache

import (
	"errors"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	store := NewCache()
	store.String().Set("key1", "value1", 10)
	store.String().Set("key2", "value2", 0) // 永不过期

	val, err := store.String().Get("key1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if val != "value1" {
		t.Fatalf("Expected value1, got %v", val)
	}

	val, err = store.String().Get("key2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if val != "value2" {
		t.Fatalf("Expected value2, got %v", val)
	}
}

func TestCache_Expire(t *testing.T) {
	store := NewCache()
	store.String().Set("key1", "value1", 0)
	store.String().Expire("key1", 1)

	time.Sleep(2 * time.Second)

	_, err := store.String().Get("key1")
	if !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("Expected ErrKeyExpired, got %v", err)
	}
}

func TestCache_Del(t *testing.T) {
	store := NewCache()
	store.String().Set("key1", "value1", 0)
	store.String().Del("key1")

	_, err := store.String().Get("key1")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestCache_Cleanup(t *testing.T) {
	store := NewCache()
	store.String().Set("key1", "value1", 1)

	time.Sleep(2 * time.Second)

	_, err := store.String().Get("key1")
	if !errors.Is(err, ErrKeyNotFound) && !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("Expected key to be expired or not found, got %v", err)
	}
}
