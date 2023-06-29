package quill_test

import (
	"testing"

	"github.com/EliCDavis/quill"
	"github.com/stretchr/testify/assert"
)

func TestPermissionTable_TryAddConflicts(t *testing.T) {
	// ARRANGE ================================================================
	table := quill.NewPermissionTable()

	// ACT / ASSERT ===========================================================

	assert.Panics(t, func() {
		table.TryAdd(map[string]quill.PermissionType{
			"something":      quill.ReadPermissionType,
			"something.else": quill.WritePermissionType,
		})
	})
}

func TestPermissionTable_Conflicts(t *testing.T) {
	// ARRANGE ================================================================
	table := quill.NewPermissionTable()

	added := table.TryAdd(map[string]quill.PermissionType{
		"something.else": quill.ReadPermissionType,
		"baseWrite":      quill.WritePermissionType,
		"baseRead":       quill.ReadPermissionType,
	})

	// ACT / ASSERT ===========================================================
	assert.Equal(t, true, added)
	assert.Equal(t, 1, table.Version())

	tests := map[string]struct {
		input     map[string]quill.PermissionType
		conflicts bool
	}{
		"empty permissions no conflict": {
			conflicts: false,
			input:     map[string]quill.PermissionType{},
		},
		"read(a.b) on read(a.b): no conflict": {
			conflicts: false,
			input: map[string]quill.PermissionType{
				"something.else": quill.ReadPermissionType,
			},
		},
		"write(a.b) on read(a.b): conflict": {
			conflicts: true,
			input: map[string]quill.PermissionType{
				"something.else": quill.WritePermissionType,
			},
		},
		"read(a.b) on write(a): conflict": {
			conflicts: true,
			input: map[string]quill.PermissionType{
				"baseWrite.sub": quill.ReadPermissionType,
			},
		},
		"write(a.b) on write(a): conflict": {
			conflicts: true,
			input: map[string]quill.PermissionType{
				"baseWrite.sub": quill.WritePermissionType,
			},
		},
		"write(a.b) on read(a): conflict": {
			conflicts: true,
			input: map[string]quill.PermissionType{
				"baseRead.sub": quill.WritePermissionType,
			},
		},
		"read(a.b) on read(a): no conflict": {
			conflicts: false,
			input: map[string]quill.PermissionType{
				"baseRead.sub": quill.ReadPermissionType,
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.conflicts, table.Conflicts(tc.input))
		})
	}
}
