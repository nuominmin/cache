package cache

import (
	"errors"
	"time"
)

type String interface {
	Set(key string, value interface{}, ttl int64) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Expire(ky string, ttl int64) error
}

func (db *memoryCache) String() String {
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

// Set 是一个泛型函数，用于设置键值对
func Set[T any](cache String, key string, value T, ttl int64) error {
	if cache == nil {
		return ErrDBClosed
	}
	return cache.Set(key, value, ttl)
}

// Get 是一个泛型函数，用于获取键值对
func Get[T any](cache String, key string) (value T, err error) {
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

// GetOrSet 获取一个键值对，如果键不存在则设置一个默认值并返回
func GetOrSet[T any](cache String, key string, valueFn func() (T, error), ttl int64) (value T, err error) {
	if value, err = Get[T](cache, key); err == nil {
		return value, nil
	}

	if errors.Is(err, ErrKeyNotFound) {
		if value, err = valueFn(); err == nil {
			err = cache.Set(key, value, ttl)
			return value, err
		}
		return value, err
	}

	return value, err
}
