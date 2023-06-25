package quill_test

import (
	"testing"

	"github.com/EliCDavis/quill"
	"github.com/stretchr/testify/assert"
)

func TestDataSourceSingleReadCommand(t *testing.T) {
	// ASSERT =================================================================
	type FloatArrView struct {
		FloatArr *quill.ArrayReadPermission[float64]
	}

	data := NastyData{
		FloatArr: []float64{1, 2, 3},
		StrArr:   []string{"1", "2", "3"},
		Sub: struct {
			IntArr []int
			Str    string
		}{
			IntArr: []int{4, 5, 6},
			Str:    "Test String",
		},
	}

	dataSource := quill.NewDataSource(data)
	sum := 0.

	// ACT ====================================================================
	dataSource.Run(&quill.ViewCommand[FloatArrView]{
		Action: func(view FloatArrView) error {
			floatData := view.FloatArr.Value()
			for i := 0; i < floatData.Len(); i++ {
				sum += floatData.At(i)
			}
			return nil
		},
	})
	dataSource.Close()

	// ASSERT =================================================================
	assert.Equal(t, 6., sum)
}

func TestDataSourceWriteReadCommand(t *testing.T) {
	// ASSERT =================================================================
	type ReadFloatArrView struct {
		FloatArr *quill.ArrayReadPermission[float64]
	}

	type WriteFloatArrView struct {
		FloatArr []float64
	}

	dataSource := quill.NewDataSource(NastyData{
		FloatArr: []float64{1, 2, 3},
		StrArr:   []string{"1", "2", "3"},
		Sub: struct {
			IntArr []int
			Str    string
		}{
			IntArr: []int{4, 5, 6},
			Str:    "Test String",
		},
	})
	sum := 0.

	// ACT ====================================================================
	dataSource.Run(
		&quill.ViewCommand[WriteFloatArrView]{
			Action: func(view WriteFloatArrView) error {
				arr := view.FloatArr
				for i := 0; i < len(arr); i++ {
					arr[i] *= 2
				}
				return nil
			},
		},
		&quill.ViewCommand[ReadFloatArrView]{
			Action: func(view ReadFloatArrView) error {
				floatData := view.FloatArr.Value()
				for i := 0; i < floatData.Len(); i++ {
					sum += floatData.At(i)
				}
				return nil
			},
		},
	)
	dataSource.Close()

	// ASSERT =================================================================
	assert.Equal(t, 12., sum)
}

func TestDataSourceReadCommandWithStructTags(t *testing.T) {
	// ASSERT =================================================================
	type SumView struct {
		DataToSum *quill.ArrayReadPermission[float64] `quill:"FloatArr"`
	}

	dataSource := quill.NewDataSource(NastyData{
		FloatArr: []float64{1, 2, 3},
		StrArr:   []string{"1", "2", "3"},
		Sub: struct {
			IntArr []int
			Str    string
		}{
			IntArr: []int{4, 5, 6},
			Str:    "Test String",
		},
	})
	sum := 0.

	// ACT ====================================================================
	dataSource.Run(
		&quill.ViewCommand[SumView]{
			Action: func(view SumView) error {
				floatData := view.DataToSum.Value()
				for i := 0; i < floatData.Len(); i++ {
					sum += floatData.At(i)
				}
				return nil
			},
		},
	)
	dataSource.Close()

	// ASSERT =================================================================
	assert.Equal(t, 6., sum)
}
