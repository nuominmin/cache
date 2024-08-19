package cache

import (
	"context"
	"fmt"
	"testing"
)

func TestHandler(t *testing.T) {
	handler := NewHandler()
	r := handler.GetRegistrar()
	_ = r.Register("hello", func(ctx context.Context, data interface{}) error {
		fmt.Println("hello", data)
		return nil
	})
	_ = r.Register("hello1", func(ctx context.Context, data interface{}) error {
		fmt.Println("hello1", data)
		return nil
	})

	type data struct {
		id   uint64
		name string
	}

}
