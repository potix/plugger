package main

import (
	"fmt"
	"time"
	"github.com/potix/plugger/plugger" 
	"github.com/potix/plugger/example/config" 
	"github.com/potix/plugger/example/event" 
	"github.com/potix/plugger/example/command" 
	"github.com/potix/plugger/example/result" 
)

func eventHandler(pluginHandle *plugger.PluginHandle,
     eventName string, eventParam interface{}, err error) (interface{}, error) {
	fmt.Println("event handler")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("event name", eventName)
	fmt.Println("event param 1", eventParam.(*event.EventParam).GetValue1())
	fmt.Println("event param 2", eventParam.(*event.EventParam).GetValue2())

	eventResult := result.NewEventResult() 
	eventResult.SetValue("handling ok")

	return eventResult, nil
}

func main() {
	configBefore := config.NewConfig()
	configBefore.SetValue1("before value1")
	configBefore.SetValue2("before value2")

	configAfter := config.NewConfig()
	configAfter.SetValue1("after value1")
	configAfter.SetValue2("after value2")

	commandParam := command.NewCommandParam()
	commandParam.SetValue1("value1")
	commandParam.SetValue2("value2")

	p := plugger.NewPlugger()

	p.SetVersionSafe(true)

	err, warn := p.Load("../plugins");
	if err != nil {
		fmt.Println(err)
		return
	}
	if warn != nil {
		fmt.Println(warn)
	}

	fmt.Println("------- build version -------")
	fmt.Println(p.GetBuildVersion())

	fmt.Println("------- get plugins -------")
	for _, v := range p.GetPluginNames() {
		fmt.Println(v)
	}

	fmt.Println("------- exists plugins -------")
	v := p.ExistsPluginNames([]string{"sample3","sample2","sample1", "sample0"})
	fmt.Println(v)

	fmt.Println("------- plugin info -------")
	pbv, fp ,err := p.GetPluginInfo("sample1")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pbv, fp)
	pbv, fp ,err= p.GetPluginInfo("sample2")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pbv, fp)

	fmt.Println("--------plugin1--------")
	time.Sleep(3 * time.Second)

	plugin1, err := p.NewPlugin("sample1", result.NewResultAsInterface,
	     result.NewCommandResultAsInterface, event.NewEventParamAsInterface)
        if err != nil {
		fmt.Println(err)
	} else {
		
		fmt.Println("c: plugin1 init")
		r, err := plugin1.Init(configBefore)
		if err != nil {
			fmt.Println("c: plugin1 init err", err.Error())
		} else {
			fmt.Println("c: plugin1 init result", r.(*result.Result).GetValue())
		}
		fmt.Println("c: plugin1 start")
		r, err = plugin1.Start()
		if err != nil {
			fmt.Println("c: plugin1 start err", err.Error())
		} else {
			fmt.Println("c: plugin1 start result", r.(*result.Result).GetValue())
		}
		fmt.Println("c: plugin1 reload")
		r, err = plugin1.Reload(configAfter)
		if err != nil {
			fmt.Println("c: plugin1 reload err", err.Error())
		} else {
			fmt.Println("c: plugin1 reload result", r.(*result.Result).GetValue())
		}
		plugin1.EventOn("hogehoge", eventHandler)
		fmt.Println("c: plugin1 command")
		r, err = plugin1.Command(commandParam)
		if err != nil {
			fmt.Println("c: plugin1 command err", err.Error())
		} else {
			fmt.Println("c: plugin1 command result", r.(*result.CommandResult).GetValue())
		}
		fmt.Println("--------event wait--------")
		time.Sleep(3 * time.Second)
		fmt.Println("c: plugin1 stop")
		r, err = plugin1.Stop()
		if err != nil {
			fmt.Println("c: plugin1 stop err", err.Error())
		} else {
			fmt.Println("c: plugin1 stop result", r.(*result.Result).GetValue())
		}
		fmt.Println("c: plugin1 fini")
		r, err = plugin1.Fini()
		if err != nil {
			fmt.Println("c: plugin1 fini err", err.Error())
		} else {
			fmt.Println("c: plugin1 fini result", r.(*result.Result).GetValue())
		}
	}
	if err := p.FreePlugin(plugin1); err != nil {
		fmt.Println(err)
	}

	fmt.Println("--------plugin2--------")
	time.Sleep(3 * time.Second)

	plugin2, err := p.NewPlugin("sample2", result.NewResultAsInterface,
	    result.NewCommandResultAsInterface, event.NewEventParamAsInterface)
	if  err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("c: plugin2 init")
		r, err := plugin2.Init(configBefore)
		if err != nil {
			fmt.Println("c: plugin2 init err", err.Error())
		} else {
			fmt.Println("c: plugin2 init result", r.(*result.Result).GetValue())
		}
		fmt.Println("c: plugin2 start")
		r, err = plugin2.Start()
		if err != nil {
			fmt.Println("c: plugin2 start err", err.Error())
		} else {
			fmt.Println("c: plugin2 start result", r.(*result.Result).GetValue())
		}
		fmt.Println("c: plugin2 reload")
		r, err = plugin2.Reload(configAfter)
		if err != nil {
			fmt.Println("c: plugin2 reload err", err.Error())
		} else {
			fmt.Println("c: plugin2 reload result", r.(*result.Result).GetValue())
		}
		plugin2.EventOn("fugafuga", eventHandler)
		fmt.Println("c: plugin2 command")
		r, err = plugin2.Command(commandParam)
		if err != nil {
			fmt.Println("c: plugin2 command err", err.Error())
		} else {
			fmt.Println("c: plugin2 command result", r.(*result.CommandResult).GetValue())
		}
		fmt.Println("--------event wait--------")
		time.Sleep(3 * time.Second)
		fmt.Println("c: plugin2 stop")
		r, err = plugin2.Stop()
		if err != nil {
			fmt.Println("c: plugin2 stop err", err.Error())
		} else {
			fmt.Println("c: plugin2 stop result", r.(*result.Result).GetValue())
		}
		fmt.Println("c: plugin2 fini")
		r, err = plugin2.Fini()
		if err != nil {
			fmt.Println("c: plugin2 fini err", err.Error())
		} else {
			fmt.Println("c: plugin2 fini result", r.(*result.Result).GetValue())
		}
	}

	p.NewPlugin("sample1", result.NewResultAsInterface,
	     result.NewCommandResultAsInterface, event.NewEventParamAsInterface)

	p.Unload("sample1")

	p.Free()
}
