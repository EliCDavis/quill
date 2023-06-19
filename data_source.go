package quill

type DataSource[T any] struct {
	data T
}

func NewDataSource[T any](data T) *DataSource[T] {
	return &DataSource[T]{data: data}
}

func (ds *DataSource[T]) Run(c Command) {
	commandData := c.data()
	PopulateView(ds.data, commandData)
	c.Run()
}
