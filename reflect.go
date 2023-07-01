package quill

import (
	"fmt"
	"reflect"
)

type postQueryOperation interface {
	apply()
}

type updateMapPostQueryOperation struct {
	mapSource, mapKey, mapVal reflect.Value
	field                     int
}

func (umqo updateMapPostQueryOperation) apply() {
	umqo.mapSource.SetMapIndex(umqo.mapKey, umqo.mapVal.Field(umqo.field))
}

func getValueByName(val reflect.Value, name string) (reflect.Value, bool) {
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if name == f.Name {
			return val.Field(i), true
		}
	}

	return reflect.Value{}, false
}

func populateViewStructsFromMap(source, view reflect.Value) []postQueryOperation {
	viewType := view.Type()

	sourceFieldKind := source.Kind()
	if sourceFieldKind != reflect.Map {
		panic(fmt.Errorf("source is not a map to process, is instead: '%s", sourceFieldKind.String()))
	}

	ops := make([]postQueryOperation, 0)

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)
		if !viewFieldValue.CanSet() {
			panic(fmt.Errorf("view contains the field (%s) that can not be assigned to. did you not pass a pointer?", structField.Name))
		}

		sourceName := structField.Name
		if altName, ok := structField.Tag.Lookup("quill"); ok {
			sourceName = altName
		}
		sourceField, mapHasKey := getMapValue(source, reflect.ValueOf(sourceName))
		sourceFieldKind := sourceField.Kind()
		viewFieldValueKind := viewFieldValue.Kind()

		// View is requesting write access to an array from the source data
		if viewFieldValueKind == reflect.Slice && sourceFieldKind == reflect.Slice {
			viewFieldValue.Set(sourceField)
			continue
		}

		if viewFieldValueKind == reflect.Slice && !mapHasKey {
			newMapVal := reflect.New(viewFieldValue.Type()).Elem()
			source.SetMapIndex(reflect.ValueOf(sourceName), newMapVal)
			viewFieldValue.Set(newMapVal)
			ops = append(ops, updateMapPostQueryOperation{
				mapSource: source,
				mapKey:    reflect.ValueOf(sourceName),
				mapVal:    view,
				field:     i,
			})
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

		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Struct {
			populateViewStructs(sourceField, viewFieldValue)
			continue
		}

		panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s and source is type %s", structField.Name, viewFieldValueKind.String(), sourceFieldKind.String()))
	}

	return ops
}

func populateViewStructs(source, view reflect.Value) []postQueryOperation {
	viewType := view.Type()

	ops := make([]postQueryOperation, 0)

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)
		if !viewFieldValue.CanSet() {
			panic(fmt.Errorf("view contains the field (%s) that can not be assigned to. did you not pass a pointer?", structField.Name))
		}

		sourceName := structField.Name
		if altName, ok := structField.Tag.Lookup("quill"); ok {
			sourceName = altName
		}
		sourceField, ok := getValueByName(source, sourceName)
		if !ok {
			panic(fmt.Errorf("source does not contain a field named: '%s' to populate view", sourceName))
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

		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Struct {
			populateViewStructs(sourceField, viewFieldValue)
			continue
		}

		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Map {
			mapOps := populateViewStructsFromMap(sourceField, viewFieldValue)
			ops = append(ops, mapOps...)
			continue
		}

		panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s and source is type %s", structField.Name, viewFieldValueKind.String(), sourceFieldKind.String()))
	}

	return ops
}

type ApplyChanges struct {
	changes []postQueryOperation
}

func (ac ApplyChanges) Apply() {
	for _, c := range ac.changes {
		c.apply()
	}
}

func PopulateView(source, view any) ApplyChanges {
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

	return ApplyChanges{
		changes: populateViewStructs(sourceValue, viewValue),
	}
}

func getMapValue(mapSource, key reflect.Value) (reflect.Value, bool) {
	iter := mapSource.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		if k.Interface() == key.Interface() {
			return v, true
		}
	}
	return reflect.ValueOf(nil), false
}

func permissionsStructFromMap(path string, source, view reflect.Value) map[string]PermissionType {
	permissions := make(map[string]PermissionType)
	viewType := view.Type()

	if source.Kind() != reflect.Map {
		panic(fmt.Errorf("source is not a map to process, is instead: '%s", source.Kind().String()))
	}

	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)
		if !viewFieldValue.CanSet() {
			panic(fmt.Errorf("view contains the field (%s) that can not be assigned to. did you not pass a pointer?", structField.Name))
		}

		mapKeyName := structField.Name
		if altName, ok := structField.Tag.Lookup("quill"); ok {
			mapKeyName = altName
		}

		sourceField, sourceContainsKey := getMapValue(source, reflect.ValueOf(mapKeyName))
		sourceFieldKind := sourceField.Kind()
		viewFieldValueKind := viewFieldValue.Kind()

		// View is requesting write access to an array from the map source data
		if viewFieldValueKind == reflect.Slice && (!sourceContainsKey || sourceFieldKind == reflect.Slice) {
			permissions[fmt.Sprintf("%s.%s", path, structField.Name)] = WritePermissionType
			continue
		}

		// View is requesting read only access for a specific data type
		if viewFieldValueKind == reflect.Pointer {
			newPtr := reflect.New(viewFieldValue.Type().Elem())
			viewFieldValue.Set(newPtr)

			i := viewFieldValue.Interface()
			perm, ok := i.(Permission)
			if !ok {
				panic(fmt.Errorf("view field '%s' is an interface but not a permission which is not allowed", structField.Name))
			}

			permissions[fmt.Sprintf("%s.%s", path, structField.Name)] = perm.Type()
			continue
		}

		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Struct {
			subPermissions := permissionsStruct(fmt.Sprintf("%s.%s", path, structField.Name), sourceField, viewFieldValue)
			for key, val := range subPermissions {
				permissions[key] = val
			}

			continue
		}

		// We want specific read/write access to a source's map
		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Map {
			subPermissions := permissionsStructFromMap(fmt.Sprintf("%s.%s", path, structField.Name), sourceField, viewFieldValue)
			for key, val := range subPermissions {
				permissions[key] = val
			}

			continue
		}

		panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s and source is type %s", structField.Name, viewFieldValueKind.String(), sourceFieldKind.String()))
	}
	return permissions
}

func permissionsStruct(path string, source, view reflect.Value) map[string]PermissionType {
	permissions := make(map[string]PermissionType)
	viewType := view.Type()
	for i := 0; i < viewType.NumField(); i++ {
		viewFieldValue := view.Field(i)
		structField := viewType.Field(i)
		if !viewFieldValue.CanSet() {
			panic(fmt.Errorf("view contains the field (%s) that can not be assigned to. did you not pass a pointer?", structField.Name))
		}

		sourceName := structField.Name
		if altName, ok := structField.Tag.Lookup("quill"); ok {
			sourceName = altName
		}
		sourceField, ok := getValueByName(source, sourceName)
		if !ok {
			panic(fmt.Errorf("source does not contain a field named: '%s' to populate view", sourceName))
		}

		sourceFieldKind := sourceField.Kind()
		viewFieldValueKind := viewFieldValue.Kind()

		// View is requesting write access to an array from the source data
		if sourceFieldKind == reflect.Slice && viewFieldValueKind == reflect.Slice {
			permissions[fmt.Sprintf("%s.%s", path, structField.Name)] = WritePermissionType
			continue
		}

		// View is requesting read only access for a specific data type
		if viewFieldValueKind == reflect.Pointer {
			newPtr := reflect.New(viewFieldValue.Type().Elem())
			viewFieldValue.Set(newPtr)

			i := viewFieldValue.Interface()
			perm, ok := i.(Permission)
			if !ok {
				panic(fmt.Errorf("view field '%s' is an interface but not a permission which is not allowed", structField.Name))
			}

			permissions[fmt.Sprintf("%s.%s", path, structField.Name)] = perm.Type()
			continue
		}

		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Struct {
			subPermissions := permissionsStruct(fmt.Sprintf("%s.%s", path, structField.Name), sourceField, viewFieldValue)
			for key, val := range subPermissions {
				permissions[key] = val
			}

			continue
		}

		// We want specific read/write access to a source's map
		if viewFieldValueKind == reflect.Struct && sourceFieldKind == reflect.Map {
			subPermissions := permissionsStructFromMap(fmt.Sprintf("%s.%s", path, structField.Name), sourceField, viewFieldValue)
			for key, val := range subPermissions {
				permissions[key] = val
			}

			continue
		}

		panic(fmt.Errorf("unimplemented scenario where view's field '%s' is type %s and source is type %s", structField.Name, viewFieldValueKind.String(), sourceFieldKind.String()))
	}
	return permissions
}

func calculatePermissions(source, view any) map[string]PermissionType {
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

	return permissionsStruct("", sourceValue, viewValue)
}
