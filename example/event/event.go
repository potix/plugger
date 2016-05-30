package event

type EventParam struct {
        Value1 string
        Value2 string
}

func (cp *EventParam) GetValue1() string {
        return cp.Value1
}

func (cp *EventParam) SetValue1(value1 string) {
        cp.Value1 = value1
}

func (cp *EventParam) GetValue2() string {
        return cp.Value2
}

func (cp *EventParam) SetValue2(value2 string) {
        cp.Value2 = value2
}

func NewEventParam() *EventParam {
        return new(EventParam)
}

func NewEventParamAsInterface() interface{} {
        return NewEventParam()
}
