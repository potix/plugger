package result

type EventResult struct {
	Value string
}

func (r *EventResult) GetValue() string {
	return r.Value
}

func (r *EventResult) SetValue(value string) {
	r.Value = value
}

func NewEventResult() *EventResult {
	return new(EventResult)
}

func NewEventResultAsInterface() interface{} {
	return NewEventResult()
}
