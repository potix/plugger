package command

type CommandParam struct {
        Value1 string
        Value2 string
}

func (cp *CommandParam) GetValue1() string {
        return cp.Value1
}

func (cp *CommandParam) SetValue1(value1 string) {
        cp.Value1 = value1
}

func (cp *CommandParam) GetValue2() string {
        return cp.Value2
}

func (cp *CommandParam) SetValue2(value2 string) {
        cp.Value2 = value2
}

func NewCommandParam() *CommandParam {
        return new(CommandParam)
}

func NewCommandParamAsInterface() interface{} {
        return NewCommandParam()
}
