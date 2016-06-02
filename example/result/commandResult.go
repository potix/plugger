package result

type CommandResult struct {
	Value string
}

func (r *CommandResult) GetValue() string {
	return r.Value
}

func (r *CommandResult) SetValue(value string) {
	r.Value = value
}

func NewCommandResult() *CommandResult {
	return new(CommandResult)
}

func NewCommandResultAsInterface() interface{} {
	return NewCommandResult()
}
