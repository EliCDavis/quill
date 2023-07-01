package quill

type Command interface {
	Run() error
	data() any
}

type ViewCommand[T any] struct {
	populatedData T
	Action        func(*T) error
}

func (vc *ViewCommand[T]) Run() error {
	return vc.Action(&vc.populatedData)
}

func (vc *ViewCommand[T]) data() any {
	return &vc.populatedData
}
