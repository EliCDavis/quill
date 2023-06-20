package quill

import (
	"fmt"
	"sync"
)

// Thread safe collection of permissions
type PermissionTable struct {
	permissions map[string]int // negative values means writing, 0 means unoccupied, positive values is the number of commands reading
	lock        sync.RWMutex
}

func NewPermissionTable() *PermissionTable {
	return &PermissionTable{
		permissions: make(map[string]int),
	}
}

// Assumes something else is utilizing the mutex to guarantee synchronization
func (pt *PermissionTable) unsafeConflict(newPermission map[string]PermissionType) bool {
	for key, newVal := range newPermission {
		if oldVal, ok := pt.permissions[key]; ok {
			if newVal == WritePermissionType || oldVal < 0 {
				return true
			}
		}
	}
	return false
}

func (pt *PermissionTable) Conflicts(newPermissions map[string]PermissionType) bool {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	return pt.unsafeConflict(newPermissions)
}

// Atomic operation
func (pt *PermissionTable) TryAdd(newPermissions map[string]PermissionType) bool {
	pt.lock.Lock()
	defer pt.lock.Unlock()

	if pt.unsafeConflict(newPermissions) {
		return false
	}

	for key, permission := range newPermissions {

		if permission == WritePermissionType {
			pt.permissions[key] = -1
			continue
		}

		// We're a read permission, increment the count
		if oldVal, ok := pt.permissions[key]; ok {
			pt.permissions[key] = oldVal + 1
		} else {
			pt.permissions[key] = 1
		}
	}
	return true
}

func (pt *PermissionTable) Clear(permissionsToClear map[string]PermissionType) {
	pt.lock.Lock()
	defer pt.lock.Unlock()

	for key, permission := range permissionsToClear {

		if permission == WritePermissionType {
			pt.permissions[key] = 0
			continue
		}

		// We're a read permission, decrement the count
		if oldVal, ok := pt.permissions[key]; ok {
			if oldVal == 0 {
				panic(fmt.Errorf("attempting to remove read permission '%s' but it's already 0", key))
			}
			pt.permissions[key] = oldVal - 1
		} else {
			panic(fmt.Errorf("attempting to remove read permission '%s' but it doesn't exist in the permissions table", key))
		}
	}
}
