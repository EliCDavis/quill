package quill

import "reflect"

type Permission interface {
	inject(reflect.Value)
	clear()
	Type() PermissionType
}

type PermissionType int

const (
	ReadPermissionType PermissionType = iota
	WritePermissionType
)
