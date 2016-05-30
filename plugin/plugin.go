package plugin

type event struct {
	eventName string
	eventParam interface{}
}

type Plugin interface {
        Init(config interface{}) (interface{}, error)
        Start() (interface{}, error)
        Stop() (interface{}, error)
        Reload(config interface{}) (interface{}, error)
        Fini() (interface{}, error)
        Command(cmdParam interface{}) (interface{}, error)
}

func SetPluginName(pluginName string) {
	pluginMgr.pluginName = pluginName
}

func SetNewPluginFunc(newPluginFunc func() Plugin) {
	pluginMgr.newPluginFunc = newPluginFunc
}

func SetNewConfigFunc(newConfigFunc func() interface{}) {
	pluginMgr.newConfigFunc = newConfigFunc
}

func SetNewCommandParamFunc(newCommandParamFunc func() interface{}) {
	pluginMgr.newCommandParamFunc = newCommandParamFunc
}

func EventEmit(plugin Plugin, eventName string, eventParam interface{}) error {
	event := &event {
		eventName: eventName,
		eventParam: eventParam,
	}
        return pluginMgr.eventEmit(plugin, event)
}
