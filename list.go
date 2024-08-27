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
	if db.closed {
		return ErrDBClosed
	}

	list := make([]interface{}, 0)
	if value, ok := db.lists.Load(key); ok {
		list = value.([]interface{})
	}
	list = append(values, list...)
	db.lists.Store(key, list)
	return nil
}

// RPush 将元素推入列表的右边（尾部）
func (db *memoryCache) RPush(key string, values ...interface{}) error {
	if db.closed {
		return ErrDBClosed
	}

	list := make([]interface{}, 0)
	if value, ok := db.lists.Load(key); ok {
		list = value.([]interface{})
	}

	list = append(list, values...)
	db.lists.Store(key, list)
	return nil
}

// LPop 从列表的左边（头部）弹出一个元素
func (db *memoryCache) LPop(key string) (interface{}, error) {
	if db.closed {
		return nil, ErrDBClosed
	}

	v, ok := db.lists.Load(key)
	if !ok {
		return nil, ErrKeyNotFound
	}

	list := v.([]interface{})
	if len(list) == 0 {
		return nil, ErrKeyNotFound
	}

	value := list[0]
	list = list[1:]
	if len(list) == 0 {
		db.lists.Delete(key)
	} else {
		db.lists.Store(key, list)
	}

	return value, nil
}

// RPop 从列表的右边（尾部）弹出一个元素
func (db *memoryCache) RPop(key string) (interface{}, error) {
	if db.closed {
		return nil, ErrDBClosed
	}

	v, ok := db.lists.Load(key)
	if !ok {
		return nil, ErrKeyNotFound
	}

	list := v.([]interface{})
	if len(list) == 0 {
		return nil, ErrKeyNotFound
	}

	value := list[len(list)-1]
	list = list[:len(list)-1]
	if len(list) == 0 {
		db.lists.Delete(key)
	} else {
		db.lists.Store(key, list)
	}

	return value, nil
}

// LRange 获取列表中指定范围的元素
func (db *memoryCache) LRange(key string, start, stop int) ([]interface{}, error) {
	if db.closed {
		return nil, ErrDBClosed
	}

	value, ok := db.lists.Load(key)
	if !ok {
		return nil, ErrKeyNotFound
	}

	list := value.([]interface{})

	if start < 0 || start >= len(list) {
		return nil, ErrStartIndexOutOfRange
	}

	if stop < 0 || stop >= len(list) {
		return nil, ErrStopIndexOutOfRange
	}

	return list[start : stop+1], nil
}
