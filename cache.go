package cache

import (
	"fmt"
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
	StringHandler() StringHandler
	Close()
}

// memoryCache 是一个内存数据存储，具体实现了 Cache 接口
type memoryCache struct {
	keyLocks sync.Map // map[string]*sync.Mutex
	mu       sync.Mutex

	data        sync.Map // map[string] item
	dataHandler sync.Map // map[string] HandlerFunc
	lists       sync.Map // map[string] []interface{}
	closed      bool
	closeCh     chan struct{}
	wg          sync.WaitGroup
}

func (db *memoryCache) acquireLock(t string, key string) {
	key = fmt.Sprintf("%s:%s", t, key)

	if value, ok := db.keyLocks.Load(key); ok {
		value.(*sync.Mutex).Lock()
		return
	}

	lock := &sync.Mutex{}
	db.keyLocks.Store(key, lock)
	lock.Lock()
}

func (db *memoryCache) releaseLock(t, key string) {
	key = fmt.Sprintf("%s:%s", t, key)

	if value, ok := db.keyLocks.Load(key); ok {
		value.(*sync.Mutex).Unlock()
	}
}

// NewCache 创建一个新的 Cache
func NewCache() Cache {
	db := &memoryCache{
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
			now := time.Now().UnixNano()
			db.data.Range(func(key, value any) bool {
				it := value.(item)
				if it.expiration > 0 && now > it.expiration {
					db.data.Delete(key)
				}
				return true
			})

		case <-db.closeCh:
			return
		}
	}
}

// Close 关闭 cache 并且清理任务
func (db *memoryCache) Close() {
	db.mu.Lock()
	db.closed = true
	close(db.closeCh)
	db.mu.Unlock()
	db.wg.Wait()

	db.data = sync.Map{}
}
