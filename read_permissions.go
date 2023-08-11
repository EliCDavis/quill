package quill

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/EliCDavis/iter"
)

// READING ====================================================================

func ReadArray[T any](collection CollectionReadPermission, path string) *iter.ArrayIterator[T] {
	return Read[*ArrayReadPermission[T]](collection, path).Value()
}

func ReadItem[T any](collection CollectionReadPermission, path string) T {
	return Read[*ItemReadPermission[T]](collection, path).Value()
}

func Read[T Permission](collection CollectionReadPermission, path string) T {
	key := path

	splitIndex := strings.Index(path, ".")
	if splitIndex != -1 {
		key = path[:splitIndex]
	}

	data, ok := collection.data[key]
	if !ok {
		panic(fmt.Errorf("collection contains no path: '%s'", key))
	}

	if splitIndex != -1 {
		v, ok := data.(CollectionReadPermission)
		if !ok {
			panic(fmt.Errorf("collection contains no collection of given type for path: '%s'", key))
		}
		return Read[T](v, path[splitIndex+1:])
	}

	v, ok := data.(T)
	if !ok {
		panic(fmt.Errorf("collection contains no permission of given type for path: '%s'", key))
	}

	return v
}

// COLLECTION =================================================================

type CollectionReadPermission struct {
	data map[string]Permission
}

func NewCollection(data map[string]Permission) CollectionReadPermission {
	return CollectionReadPermission{
		data: data,
	}
}

func (rcp CollectionReadPermission) Data() any {
	return rcp.data
}

func (rcp CollectionReadPermission) Populate(newData any) {
	rcp.inject(reflect.ValueOf(newData))
}

func (rcp CollectionReadPermission) inject(val reflect.Value) {
	kind := val.Kind()
	if kind == reflect.Pointer {
		panic("collections can not be populated with pointers yet")
	}

	if kind != reflect.Struct {
		panic(fmt.Errorf("collections can not be populated by %s", kind.String()))
	}

	for key, perm := range rcp.data {
		field, ok := getValueByName(val, key)
		if !ok {
			panic(fmt.Errorf("struct does not contain a field named: '%s' to populate collection", key))
		}
		perm.inject(field)
	}
}

func (rcp CollectionReadPermission) clear() {
	for _, perm := range rcp.data {
		perm.clear()
	}
}

func (rcp CollectionReadPermission) Type() PermissionType {
	return ReadPermissionType
}

// ARRAY ======================================================================

type ArrayReadPermission[T any] struct {
	data []T
}

func (rdep ArrayReadPermission[T]) Value() *iter.ArrayIterator[T] {
	return iter.Array[T](rdep.data)
}

func (rdep *ArrayReadPermission[T]) inject(val reflect.Value) {
	t := val.Kind()
	if t != reflect.Slice {
		panic(fmt.Errorf("can not populate an array permission with value of type: %s", t.String()))
	}

	rdep.data = val.Interface().([]T)
}

func (rdep *ArrayReadPermission[T]) clear() {
	rdep.data = nil
}

func (rdep ArrayReadPermission[T]) Type() PermissionType {
	return ReadPermissionType
}

// ITEM =======================================================================

type ItemReadPermission[T any] struct {
	data T
}

func (itp ItemReadPermission[T]) Value() T {
	return itp.data
}

func (itp *ItemReadPermission[T]) inject(val reflect.Value) {
	itp.data = val.Interface().(T)
}

func (itp *ItemReadPermission[T]) clear() {
	var data T
	itp.data = data
}

func (itp ItemReadPermission[T]) Type() PermissionType {
	return ReadPermissionType
}
