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

type sample2 struct {
}

func (s *sample2) Init(conf interface{}) (interface{}, error) {
        fmt.Println("p: sample2 init")
        fmt.Println("p: sample2 config 1", conf.(*config.Config).GetValue1())
        fmt.Println("p: sample2 config 2", conf.(*config.Config).GetValue2())
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample2) Start() (interface{}, error)  {
        fmt.Println("p: sample2 start")
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample2) Stop() (interface{}, error)  {
        fmt.Println("p: sample2 stop")
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample2) Reload(newConf interface{}) (interface{}, error)  {
        fmt.Println("p: sample2 reload")
        fmt.Println("p: sample2 new config 1", newConf.(*config.Config).GetValue1())
        fmt.Println("p: sample2 new config 2", newConf.(*config.Config).GetValue2())
        return nil, errors.New("not support")
}

func (s *sample2) Fini() (interface{}, error) {
        fmt.Println("p: sample2 fini")
	r := result.NewResult()
	r.SetValue("OK")
        return r, nil
}

func (s *sample2) Command(cmdParam interface{}) (interface{}, error) {
        fmt.Println("p: sample2 command")
        fmt.Println("p: sample2 command param 1", cmdParam.(*command.CommandParam).GetValue1())
        fmt.Println("p: sample2 command param 2", cmdParam.(*command.CommandParam).GetValue2())
	r := result.NewCommandResult()
	r.SetValue("OKOK")

        fmt.Println("p: sample2 event emit")
        event := event.NewEventParam()
        event.SetValue1("1111")
        event.SetValue2("2222")

        eventResult, err := plugin.EventEmit(s, event)
        if err != nil {
                println("p: event result error", err.Error())
        }
        println("p: event result", eventResult.(*result.EventResult).GetValue())

        return r, nil
}

func NewSample2() plugin.Plugin {
	return new(sample2)
}

func init() {
	plugin.SetPluginName("sample2")
	plugin.SetNewPluginFunc(NewSample2)
	plugin.SetNewConfigFunc(config.NewConfigAsInterface)
	plugin.SetNewCommandParamFunc(command.NewCommandParamAsInterface)
	plugin.SetNewEventResultFunc(result.NewEventResultAsInterface)
}

func main() {
}
