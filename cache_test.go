package cache

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	store := NewCache()
	key := "key"

	var i int
	for {
		for j := 0; j < 5; j++ {
			go func(i, j int) {
				value, _ := GetOrSet[string](store, key, func() (string, error) {
					time.Sleep(time.Second * time.Duration(5))
					return fmt.Sprintf("value,%d, %d", j, i), nil
				}, 10)
				fmt.Println(value)
			}(i, j)
		}

		time.Sleep(time.Second)
		i++
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
