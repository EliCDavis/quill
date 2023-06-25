# Quill
![Coverage](https://img.shields.io/badge/Coverage-69.5%25-yellow)

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

## Profiling

The data source uses `runtime/trace` to help track how well operations are getting parallelized over it.

```
go test . -trace trace.out
go tool trace trace.out
```
