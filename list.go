package cache

type List interface {
	LPush(key string, values ...interface{}) error
	RPush(key string, values ...interface{}) error
	LPop(key string) (interface{}, error)
	RPop(key string) (interface{}, error)
	LRange(key string, start, stop int) ([]interface{}, error)
}

func (db *memoryCache) List() List {
	return db
}

// LPush 将元素推入列表的左边（头部）
func (db *memoryCache) LPush(key string, values ...interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	list, ok := db.lists[key]
	if !ok {
		list = make([]interface{}, 0)
	}

	list = append(values, list...)
	db.lists[key] = list
	return nil
}

// RPush 将元素推入列表的右边（尾部）
func (db *memoryCache) RPush(key string, values ...interface{}) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return ErrDBClosed
	}

	list, ok := db.lists[key]
	if !ok {
		list = make([]interface{}, 0)
	}

	list = append(list, values...)
	db.lists[key] = list
	return nil
}

// LPop 从列表的左边（头部）弹出一个元素
func (db *memoryCache) LPop(key string) (interface{}, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return nil, ErrDBClosed
	}

	list, ok := db.lists[key]
	if !ok || len(list) == 0 {
		return nil, ErrKeyNotFound
	}

	value := list[0]
	list = list[1:]
	if len(list) == 0 {
		delete(db.lists, key)
	} else {
		db.lists[key] = list
	}

	return value, nil
}

// RPop 从列表的右边（尾部）弹出一个元素
func (db *memoryCache) RPop(key string) (interface{}, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.closed {
		return nil, ErrDBClosed
	}

	list, ok := db.lists[key]
	if !ok || len(list) == 0 {
		return nil, ErrKeyNotFound
	}

	value := list[len(list)-1]
	list = list[:len(list)-1]
	if len(list) == 0 {
		delete(db.lists, key)
	} else {
		db.lists[key] = list
	}

	return value, nil
}

// LRange 获取列表中指定范围的元素
func (db *memoryCache) LRange(key string, start, stop int) ([]interface{}, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if db.closed {
		return nil, ErrDBClosed
	}

	list, ok := db.lists[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if start < 0 || start >= len(list) {
		return nil, ErrStartIndexOutOfRange
	}

	if stop < 0 || stop >= len(list) {
		return nil, ErrStopIndexOutOfRange
	}

	return list[start : stop+1], nil
}
