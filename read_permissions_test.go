package quill_test

import (
	"testing"

	"github.com/EliCDavis/quill"
	"github.com/stretchr/testify/assert"
)

// func TestReadCollection(t *testing.T) {
// 	// ASSERT =================================================================
// 	collection := quill.NewCollection(map[string]quill.Permission{
// 		"FloatArr": quill.ArrayReadPermission[float64],
// 		"StrArr":   &quill.ArrayWritePermission[string]{},
// 		"Sub": quill.NewCollection(map[string]quill.Permission{
// 			"IntArr": &quill.ArrayReadPermission[int]{},
// 			"Str":    &quill.ItemReadPermission[string]{},
// 		}),
// 	})

// 	data := struct {
// 		FloatArr []float64
// 		StrArr   []string
// 		Sub      struct {
// 			IntArr []int
// 			Str    string
// 		}
// 	}{
// 		FloatArr: []float64{1, 2, 3},
// 		StrArr:   []string{"1", "2", "3"},
// 		Sub: struct {
// 			IntArr []int
// 			Str    string
// 		}{
// 			IntArr: []int{4, 5, 6},
// 			Str:    "Test String",
// 		},
// 	}

// 	// ACT ====================================================================
// 	collection.Populate(data)
// 	floatData := quill.ReadArray[float64](collection, "FloatArr")
// 	intData := quill.ReadArray[int](collection, "Sub/IntArr")
// 	strData := quill.ReadItem[string](collection, "Sub/Str")

// 	// ASSERT =================================================================
// 	if assert.Equal(t, floatData.Len(), 3) {
// 		assert.Equal(t, 1., floatData.At(0))
// 		assert.Equal(t, 2., floatData.At(1))
// 		assert.Equal(t, 3., floatData.At(2))
// 	}

// 	if assert.Equal(t, intData.Len(), 3) {
// 		assert.Equal(t, 4, intData.At(0))
// 		assert.Equal(t, 5, intData.At(1))
// 		assert.Equal(t, 6, intData.At(2))
// 	}

// 	assert.Equal(t, "Test String", strData)
// }

func TestView(t *testing.T) {
	// ASSERT =================================================================
	view := struct {
		FloatArr []float64
		StrArr   *quill.ArrayReadPermission[string]
	}{}

	data := struct {
		FloatArr []float64
		StrArr   []string
		Sub      struct {
			IntArr []int
			Str    string
		}
	}{
		FloatArr: []float64{-1, -2, -3},
		StrArr:   []string{"1", "2", "3"},
		Sub: struct {
			IntArr []int
			Str    string
		}{
			IntArr: []int{4, 5, 6},
			Str:    "Test String",
		},
	}

	// ACT ====================================================================
	quill.PopulateView(data, &view)

	// ASSERT =================================================================
	strView := view.StrArr.Value()
	if assert.Equal(t, 3, strView.Len()) {
		assert.Equal(t, "1", strView.At(0))
		assert.Equal(t, "2", strView.At(1))
		assert.Equal(t, "3", strView.At(2))
	}

	if assert.Equal(t, 3, len(view.FloatArr)) {
		assert.Equal(t, -1., view.FloatArr[0])
		assert.Equal(t, -2., view.FloatArr[1])
		assert.Equal(t, -3., view.FloatArr[2])
	}
}
