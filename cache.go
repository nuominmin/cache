package cache

import (
	"sync"
	"time"
)

type item struct {
	object     interface{}
	expiration int64
}

// Cache 缓存接口
type Cache interface {
	List() List
	String() String
	Close()
}

// memoryCache 是一个内存数据存储，具体实现了 Cache 接口
type memoryCache struct {
	data    map[string]item
	lists   map[string][]interface{}
	mutex   sync.RWMutex
	closed  bool
	closeCh chan struct{}
	wg      sync.WaitGroup
}

// NewCache 创建一个新的 Cache
func NewCache() Cache {
	db := &memoryCache{
		data:    make(map[string]item),
		lists:   make(map[string][]interface{}),
		closeCh: make(chan struct{}),
	}
	db.wg.Add(1)
	go db.cleanupExpiredKeys()
	return db
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
