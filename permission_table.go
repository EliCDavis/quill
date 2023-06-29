package quill

import (
	"fmt"
	"strings"
	"sync"
)

type permissionLayer struct {
	// negative values means writing, 0 means unoccupied, positive values is
	// the number of commands reading.
	permissions map[string]int
	children    map[string]*permissionLayer
}

func newPermissionLayer() permissionLayer {
	return permissionLayer{
		permissions: make(map[string]int),
		children:    make(map[string]*permissionLayer),
	}
}

func (pl *permissionLayer) Conflict(keys []string, newPerm PermissionType) bool {
	if len(keys) == 0 {
		panic("conflict should never be passed 0 keys")
	}

	rootKey := keys[0]
	if curPerm, ok := pl.permissions[rootKey]; ok {
		if newPerm == WritePermissionType || curPerm < 0 {
			return true
		}
	}

	if len(keys) > 1 {
		if curPerm, ok := pl.children[rootKey]; ok {
			return curPerm.Conflict(keys[1:], newPerm)
		}
	}

	return false
}

func (pl *permissionLayer) Add(keys []string, newPerm PermissionType) {
	if len(keys) == 0 {
		panic("add should never be passed 0 keys")
	}

	rootKey := keys[0]

	if len(keys) == 1 {
		switch newPerm {

		case WritePermissionType:
			pl.permissions[rootKey] = -1

		case ReadPermissionType:
			if oldVal, ok := pl.permissions[rootKey]; ok {
				pl.permissions[rootKey] = oldVal + 1
			} else {
				pl.permissions[rootKey] = 1
			}
		}

		return
	}

	if _, ok := pl.children[rootKey]; !ok {
		layer := newPermissionLayer()
		pl.children[rootKey] = &layer
	}
	pl.children[rootKey].Add(keys[1:], newPerm)
}

func (pl *permissionLayer) Clear(keys []string) {
	if len(keys) == 0 {
		panic("add should never be passed 0 keys")
	}

	rootKey := keys[0]

	if len(keys) == 1 {
		if perm, ok := pl.permissions[rootKey]; ok {
			if perm < 0 {
				pl.permissions[rootKey] = 0
			} else if perm > 0 {
				pl.permissions[rootKey] = perm - 1
			} else {
				panic(fmt.Errorf("trying to clear permission %s that's already clear", rootKey))
			}
		} else {
			panic(fmt.Errorf("trying to clear permission %s that's never been set", rootKey))
		}
		return
	}

	if layer, ok := pl.children[rootKey]; ok {
		layer.Clear(keys[1:])
	} else {
		panic(fmt.Errorf("trying to clear permission %s that's never been set", keys))
	}
}

// Thread safe collection of permissions
type PermissionTable struct {
	permissions permissionLayer
	changes     int
	lock        sync.RWMutex
}

func NewPermissionTable() *PermissionTable {
	return &PermissionTable{
		permissions: newPermissionLayer(),
	}
}

// Assumes something else is utilizing the mutex to guarantee synchronization
func (pt *PermissionTable) unsafeConflict(newPermission map[string]PermissionType) bool {
	for key, newVal := range newPermission {
		path := strings.Split(key, ".")
		if pt.permissions.Conflict(path, newVal) {
			return true
		}
	}
	return false
}

// Atomic operation
func (pt *PermissionTable) Conflicts(newPermissions map[string]PermissionType) bool {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	return pt.unsafeConflict(newPermissions)
}

// Atomic operation
func (pt *PermissionTable) Version() int {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	return pt.changes
}

// Atomic operation
func (pt *PermissionTable) TryAdd(newPermissions map[string]PermissionType) bool {
	pt.lock.Lock()
	defer pt.lock.Unlock()

	if pt.unsafeConflict(newPermissions) {
		return false
	}

	pt.changes++

	for key, permission := range newPermissions {
		keys := strings.Split(key, ".")
		if pt.permissions.Conflict(keys, permission) {
			panic(fmt.Errorf("%s permission conflicts with the rest of the block attempting to be added", key))
		}
		pt.permissions.Add(keys, permission)
	}
	return true
}

func (pt *PermissionTable) Clear(permissionsToClear map[string]PermissionType) {
	pt.lock.Lock()
	defer pt.lock.Unlock()

	pt.changes++

	for key := range permissionsToClear {
		pt.permissions.Clear(strings.Split(key, "."))
	}
}
