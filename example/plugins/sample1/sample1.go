package main

import (
    "fmt"
    "errors"
    "github.com/potix/plugger/plugin"
    "github.com/potix/plugger/example/config"
    "github.com/potix/plugger/example/event"
    "github.com/potix/plugger/example/command"
    "github.com/potix/plugger/example/result"
)

type sample1 struct {
}

func (s *sample1) Init(conf interface{}) (interface{}, error) {
        fmt.Println("p: sample1 init")
        fmt.Println("p: sample1 config 1", conf.(*config.Config).GetValue1())
        fmt.Println("p: sample1 config 2", conf.(*config.Config).GetValue2())
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample1) Start() (interface{}, error)  {
        fmt.Println("p: sample1 start")
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample1) Stop() (interface{}, error)  {
        fmt.Println("p: sample1 stop")
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample1) Reload(newConf interface{}) (interface{}, error)  {
        fmt.Println("p: sample1 reload")
        fmt.Println("p: sample1 new config 1", newConf.(*config.Config).GetValue1())
        fmt.Println("p: sample1 new config 2", newConf.(*config.Config).GetValue2())
        return nil, errors.New("not support")
}

func (s *sample1) Fini() (interface{}, error) {
        fmt.Println("p: sample1 fini")
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample1) Command(cmdParam interface{}) (interface{}, error) {
        fmt.Println("p: sample1 command")
        fmt.Println("p: sample1 command param 1", cmdParam.(*command.CommandParam).GetValue1())
        fmt.Println("p: sample1 command param 2", cmdParam.(*command.CommandParam).GetValue2())
	r := result.NewCommandResult()
	r.SetValue("OKOK")

	fmt.Println("p: sample1 event emit")
	event := event.NewEventParam()
	event.SetValue1("1111")
	event.SetValue2("2222")

	eventResult ,err := plugin.EventEmit(s, event)
	if err != nil {
		println("p: event result error", err.Error())
	}
	println("p: event result", eventResult.(*result.EventResult).GetValue())

        return r, nil
}

func NewSample1() plugin.Plugin {
	return new(sample1)
}

func init() {
	plugin.SetPluginName("sample1")
	plugin.SetNewPluginFunc(NewSample1)
	plugin.SetNewConfigFunc(config.NewConfigAsInterface)
	plugin.SetNewCommandParamFunc(command.NewCommandParamAsInterface)
	plugin.SetNewEventResultFunc(result.NewEventResultAsInterface)
}

func main() {
}
