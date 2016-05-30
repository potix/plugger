package result

type Result struct {
	Value string
}

func (r *Result) GetValue() string {
	return r.Value
}

func (r *Result) SetValue(value string) {
	r.Value = value
}

func NewResult() *Result {
	return new(Result)
}

func NewResultAsInterface() interface{} {
	return NewResult()
}
