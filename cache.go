package cache

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyNotFound       = errors.New("key not found")
	ErrKeyExpired        = errors.New("key expired")
	ErrDBClosed          = errors.New("cache is closed")
	ErrTypeAssertionFail = errors.New("type assertion failed")
)

type item struct {
	object     interface{}
	expiration int64
}

// Cache 缓存接口
type Cache interface {
	Set(key string, value interface{}, ttl int64) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Expire(ky string, ttl int64) error
	Close()
}

// Set 是一个泛型函数，用于设置键值对
func Set[T any](cache Cache, key string, value T, ttl int64) error {
	if cache == nil {
		return ErrDBClosed
	}
	return cache.Set(key, value, ttl)
}

// Get 是一个泛型函数，用于获取键值对
func Get[T any](cache Cache, key string) (value T, err error) {
	if cache == nil {
		return value, ErrDBClosed
	}
	var v interface{}
	if v, err = cache.Get(key); err != nil {
		return value, err
	}
	var ok bool
	if value, ok = v.(T); !ok {
		return value, ErrTypeAssertionFail
	}
	return value, nil
}

// memoryCache 是一个内存数据存储，具体实现了 Cache 接口
type memoryCache struct {
	data    map[string]item
	mutex   sync.RWMutex
	closed  bool
	closeCh chan struct{}
	wg      sync.WaitGroup
}

// NewCache 创建一个新的 Cache
func NewCache() Cache {
	db := &memoryCache{
		data:    make(map[string]item),
		closeCh: make(chan struct{}),
	}
	db.wg.Add(1)
	go db.cleanupExpiredKeys()
	return db
}

// Set 设置一个键值对，可以选择设置过期时间（以秒为单位）
func (db *memoryCache) Set(key string, value interface{}, ttl int64) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	expiration := int64(0)
	if ttl > 0 {
		expiration = time.Now().UnixNano() + ttl*int64(time.Second)
	}

	db.data[key] = item{
		object:     value,
		expiration: expiration,
	}
	return nil
}

// Get 获取一个键的值，如果键不存在或者已过期则返回错误
func (db *memoryCache) Get(key string) (interface{}, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if db.closed {
		return nil, ErrDBClosed
	}

	it, found := db.data[key]
	if !found {
		return nil, ErrKeyNotFound
	}

	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		delete(db.data, key)
		return nil, ErrKeyExpired
	}

	return it.object, nil
}

// Del 删除一个键
func (db *memoryCache) Del(key string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	delete(db.data, key)
	return nil
}

// Expire 设置一个键的过期时间（以秒为单位）
func (db *memoryCache) Expire(key string, ttl int64) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	it, ok := db.data[key]
	if !ok {
		return ErrKeyNotFound
	}

	if ttl > 0 {
		it.expiration = time.Now().UnixNano() + ttl*int64(time.Second)
	} else {
		it.expiration = 0
	}

	db.data[key] = it
	return nil
}

// cleanupExpiredKeys 定期清理过期的键
func (db *memoryCache) cleanupExpiredKeys() {
	defer db.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			db.mutex.Lock()
			if len(db.data) > 0 {
				now := time.Now().UnixNano()
				for k, it := range db.data {
					if it.expiration > 0 && now > it.expiration {
						delete(db.data, k)
					}
				}
			}
			db.mutex.Unlock()
		case <-db.closeCh:
			return
		}
	}
}

// Close 关闭 cache 并且清理任务
func (db *memoryCache) Close() {
	db.mutex.Lock()
	db.closed = true
	close(db.closeCh)
	db.mutex.Unlock()
	db.wg.Wait()

	db.mutex.Lock()
	db.data = make(map[string]item)
	db.mutex.Unlock()
}
