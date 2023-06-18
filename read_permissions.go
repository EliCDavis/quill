package quill

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/EliCDavis/iter"
)

// READING ====================================================================

func populateIterator(source, view reflect.Value) {
	// sourceElement := source.Elem()
	// sourceElementType := sourceElement.Type()

	newIter := iter.Array(source.Interface().([]string))
	view.Set(reflect.ValueOf(newIter))
}

func PopulateView(source, view any) {
	sourceValue := reflect.ValueOf(source)
	sourceKind := sourceValue.Kind()
	if sourceKind == reflect.Pointer {
		panic("populating a view with a pointer to a source is not supported yet")
	}

	if sourceKind != reflect.Struct {
		panic(fmt.Errorf("views can not be populated by sources of type: %s", sourceKind.String()))
	}

	viewPointerValue := reflect.ValueOf(view)
	viewPointerKind := viewPointerValue.Kind()
	if viewPointerKind != reflect.Pointer {
		panic("populating a view with a non-pointer to a source is not supported")
	}

	viewValue := viewPointerValue.Elem()
	viewKind := viewValue.Kind()
	if viewKind != reflect.Struct {
		panic(fmt.Errorf("views of type: '%s' can not be populated", viewKind.String()))
	}

	viewType := viewValue.Type()
	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := viewValue.Field(i)
		structField := viewType.Field(i)
		if !viewFieldValue.CanSet() {
			panic(fmt.Errorf("view contains the field (%s) that can not be assigned to. did you not pass a pointer?", structField.Name))
		}

		sourceField, ok := getValueByName(sourceValue, structField.Name)
		if !ok {
			panic(fmt.Errorf("source does not contain a field named: '%s' to populate view", structField.Name))
		}

		sourceFieldKind := sourceField.Kind()
		viewFieldValueKind := viewFieldValue.Kind()

		// View is requesting write access to an array from the source data
		if sourceFieldKind == reflect.Slice && viewFieldValueKind == reflect.Slice {
			viewFieldValue.Set(sourceField)
			continue
		}

		// View is requesting read only access
		if viewFieldValueKind == reflect.Pointer {
			newPtr := reflect.New(viewFieldValue.Type().Elem())
			viewFieldValue.Set(newPtr)

			i := viewFieldValue.Interface()
			perm, ok := i.(Permission)
			if !ok {
				panic(fmt.Errorf("view field '%s' is an interface but not a permission which is not allowed", structField.Name))
			}

			perm.inject(sourceField)
			continue
		}
	}
}

func ReadArray[T any](collection CollectionReadPermission, path string) iter.ArrayIterator[T] {
	return Read[*ArrayReadPermission[T]](collection, path).Value()
}

func ReadItem[T any](collection CollectionReadPermission, path string) T {
	return Read[*ItemReadPermission[T]](collection, path).Value()
}

func Read[T Permission](collection CollectionReadPermission, path string) T {
	key := path

	splitIndex := strings.Index(path, "/")
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

// ARRAY ======================================================================

// type ArrayReadPermission[T any] interface {
// 	inject(val reflect.Value)
// 	clear()
// 	Value() iter.ArrayIterator[T]
// }

type ArrayReadPermission[T any] struct {
	data []T
}

func (rdep ArrayReadPermission[T]) Value() iter.ArrayIterator[T] {
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
