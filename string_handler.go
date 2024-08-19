package cache

import (
	"context"
	"errors"
	"sync"
)

type HandlerFunc func(ctx context.Context) (value interface{}, ttl int64, err error)

// Handler 是一个管理处理函数的结构
type Handler interface {
	RegisterHandler(key string, handlerFunc HandlerFunc) error
	GetOrSet(ctx context.Context, key string) (interface{}, error)
}

// HandlerManage 是一个管理处理函数的结构
type HandlerManage struct {
	handler map[string]HandlerFunc
	mutex   sync.RWMutex
}

// StringHandler 创建一个新的 Handler
func (db *memoryCache) StringHandler() Handler {
	return db
}

// RegisterHandler 添加一个处理函数，并确保 key 唯一
func (db *memoryCache) RegisterHandler(key string, handlerFunc HandlerFunc) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.dataHandler[key]; ok {
		return ErrHandlerKeyExists
	}

	db.dataHandler[key] = handlerFunc
	return nil
}

// GetOrSet 获取一个键值对，如果键不存在则设置一个默认值并返回
func (db *memoryCache) GetOrSet(ctx context.Context, key string) (value interface{}, err error) {
	if value, err = db.Get(key); err == nil {
		return value, nil
	}
	if !errors.Is(err, ErrKeyNotFound) {
		return nil, err
	}

	handler, ok := db.dataHandler[key]
	if !ok {
		return nil, ErrHandlerKeyNotFound
	}

	var llt int64
	if value, llt, err = handler(ctx); err != nil {
		return nil, err
	}

	return value, db.Set(key, value, llt)
}
