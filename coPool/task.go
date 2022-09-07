package coPool

type task struct {
	param      interface{}
	handleFunc func(interface{}) error
}

func NewTask(data interface{}, f func(interface{}) error) *task {
	t := &task{
		param:      data,
		handleFunc: f,
	}
	return t
}

func (t *task) Execute() error {
	return t.handleFunc(t.param)
}
