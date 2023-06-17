package quill

import "reflect"

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
