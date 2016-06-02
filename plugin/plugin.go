package plugin

type Plugin interface {
        Init(config interface{}) (result interface{}, err error)
        Start() (result interface{}, err error)
        Stop() (result interface{}, err error)
        Reload(config interface{}) (result interface{}, err error)
        Fini() (result interface{}, err error)
        Command(cmdParam interface{}) (commandResult interface{}, err error)
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

func SetNewEventResultFunc(newEventResultFunc func() interface{}) {
	pluginMgr.newEventResultFunc = newEventResultFunc
}

func EventEmit(plugin Plugin, eventName string, eventParam interface{}) (eventResult interface{}, handleErr error) {
	eventRequest := &eventRequest {
		eventName: eventName,
		eventParam: eventParam,
		eventResChan : make(chan *eventResponse),
	}
        return pluginMgr.eventEmit(plugin, eventRequest)
}
