package quill

import (
	"fmt"
	"reflect"
)

type ArrayWritePermission[T any] struct {
	data []T
}

func (awp ArrayWritePermission[T]) Value() []T {
	return awp.data
}

func (awp *ArrayWritePermission[T]) inject(val reflect.Value) {
	t := val.Kind()
	if t != reflect.Slice {
		panic(fmt.Errorf("can not populate an array permission with value of type: %s", t.String()))
	}

	awp.data = val.Interface().([]T)
}

func (awp *ArrayWritePermission[T]) clear() {
	awp.data = nil
}

type WritePermission[T any] struct {
	data    T
	written bool
}

func (wp WritePermission[T]) Data() T {
	return wp.data
}

func (wp *WritePermission[T]) Write(val T) {
	wp.data = val
	wp.written = true
}
