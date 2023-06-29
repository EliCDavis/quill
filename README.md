# Quill

![Coverage](https://img.shields.io/badge/Coverage-71.2%25-brightgreen)

Scheduler of operations on in-memory data. The rabbit hole has gone to far.

Designed for parallelizing expensive operations on any arbitrary data collection by taking advantage of explicit read/write declaration.

## Explanation

Imagine your nasty data as a traditional database. You need to make some queries on it. You need a bunch of different operations to run on this database. The database is actually all in-memory though. How do you design your system to allow as much possible parallel execution against your data as possible?

For sake of example, let's say our **source** is defined as:

```golang
type NastyData struct {
    FloatArr []float64
    StrArr   []string
    Sub      struct {
        IntArr []int
        Str    string
    }
}
```

Let's say we wanted to sum all the data found in the `FloatArr` field. To do that, we only need read only access to that specific field in the data-structure and nothing else. To do that, we would declare our view as:

```golang
type FloatView struct {
    FloatArr *quill.ArrayReadPermission[float64]
}
```

Alternatively, if you don't want to match the view's struct field name to the source's struct field name, you can use tags:

```golang
type FloatView struct {
    DataToSum *quill.ArrayReadPermission[float64] `quill:"FloatArr"`
}
```

And then to actually perform our query:

```golang
// Create a new data source that wraps our data structure we want to
// parallelize queries against and initialize with dummy data for sake of the
// example
dataSource := quill.NewDataSource(NastyData{
    FloatArr: []float64{1, 2, 3},
})

sum := 0.

dataSource.Run(&quill.ViewCommand[FloatView]{
    Action: func(view FloatView) error {
        // All read-only array data is wrapped in an iterator to prevent us
        // from making any changes to it
        floatData := view.FloatArr.Value()
        for i := 0; i < floatData.Len(); i++ {
            sum += floatData.At(i)
        }
        return nil
    },
})
dataSource.Wait()

log.Print(sum) // prints '6'
```

### Maps

You can also request specific read/write access to entries of maps found within source data. Given our source data looks something like:

```golang
type CSVData struct {
    Title   string
    Columns map[string][]float64
}
```

We can index specific columns of the CSV data simply by defining them in the view struct

```golang
type TaxCalculationView struct {
    Columns struct {
        BasePrice  *quill.ArrayReadPermission[float64]
        TaxRate    *quill.ArrayReadPermission[float64]
        TotalPrice []float64                            // Write access to Total Price Column
    }
}
```

And then running our commands over our source data looks practically the same.

```golang
dataSource := quill.NewDataSource(CSVData{
    Columns: map[string][]float64{
        "BasePrice":  []float64{10.,   12.,  22.},
        "TaxRate":    []float64{0.07, 0.08, 0.09},
        "TotalPrice": []float64{},
    }
})

dataSource.Run(&quill.ViewCommand[TaxCalculationView]{
    Action: func(view TaxCalculationView) error {
        basePrice := view.Columns.BasePrice.Value()
        taxRate := view.Columns.TaxRate.Value()

        if basePrice.Len() != taxRate.Len() {
            return errors.New("mismatched column lengths")
        }

        totalPrice := view.Columns.TotalPrice
        for i := 0; i < basePrice.Len(); i++ {
            initialPrice := basePrice.At(i)
            totalPrice[i] = initialPrice + (initialPrice * taxRate.At(i))
        }
        return nil
    },
})
dataSource.Wait()
```

## Profiling

The data source uses `runtime/trace` to help track how well operations are getting parallelized over it.

```
go test . -trace trace.out
go tool trace trace.out
```
