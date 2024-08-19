package cache

import "errors"

var (
	ErrKeyNotFound          = errors.New("key not found")
	ErrKeyExpired           = errors.New("key expired")
	ErrDBClosed             = errors.New("cache is closed")
	ErrTypeAssertionFail    = errors.New("type assertion failed")
	ErrStartIndexOutOfRange = errors.New("start index out of range")
	ErrStopIndexOutOfRange  = errors.New("stop index out of range")
	ErrHandlerKeyExists     = errors.New("handler key already exists")
	ErrHandlerKeyNotFound   = errors.New("handler key not found")
)
