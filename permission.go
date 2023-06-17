package quill

import "reflect"

type Permission interface {
	inject(reflect.Value)
	clear()
}
