package cache

import (
	"context"
	"errors"
	"sync"
)

type HandlerFunc func(ctx context.Context) (value interface{}, ttl int64, err error)

// StringHandler 是一个管理处理函数的结构
type StringHandler interface {
	RegisterHandler(key string, handlerFunc HandlerFunc) error
	GetOrSet(ctx context.Context, key string) (interface{}, error)
}

// HandlerManage 是一个管理处理函数的结构
type HandlerManage struct {
	handler map[string]HandlerFunc
	mutex   sync.RWMutex
}

// StringHandler 创建一个新的 Handler
func (db *memoryCache) StringHandler() StringHandler {
	return db
}

// RegisterHandler 添加一个处理函数，并确保 key 唯一
func (db *memoryCache) RegisterHandler(key string, handlerFunc HandlerFunc) error {
	if _, ok := db.dataHandler.Load(key); ok {
		return ErrHandlerKeyExists
	}

	db.dataHandler.Store(key, handlerFunc)
	return nil
}

// GetOrSet 获取一个键值对，如果键不存在则设置一个默认值并返回
func (db *memoryCache) GetOrSet(ctx context.Context, key string) (value interface{}, err error) {
	db.acquireLock("String", key)
	defer db.releaseLock("String", key)

	if value, err = db.Get(key); err == nil {
		return value, nil
	}
	if !errors.Is(err, ErrKeyNotFound) {
		return nil, err
	}

	v, ok := db.dataHandler.Load(key)
	if !ok {
		return nil, ErrHandlerKeyNotFound
	}

	handler := v.(HandlerFunc)

	var llt int64
	if value, llt, err = handler(ctx); err != nil {
		return nil, err
	}

	return value, db.Set(key, value, llt)
}
