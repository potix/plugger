package plugger

import (
	"sync"
	"unsafe"
	"github.com/potix/plugger/common"
)

const (
        smallBufferSize = 512
        bigBufferSize = 2048
)

type bufferManager struct {
	bufferSize int
}

func (b *bufferManager) getBuffer() []byte {
	return make([]byte, 0, b.bufferSize)
}

func (b *bufferManager) glowBuffer(oldBuffer []byte) []byte {
	b.bufferSize = cap(oldBuffer) * 2
	return make([]byte, 0, b.bufferSize)
}

func newBufferManager(initBufferSize int) *bufferManager {
	return &bufferManager {
		bufferSize : initBufferSize,
	}	
}

type dynLoadLib struct {
        pluginName        string
        dlHandle          *dlHandle
        getBuildVersion   unsafe.Pointer
        getName           unsafe.Pointer
        newPlugin         unsafe.Pointer
        initPlugin        unsafe.Pointer
        startPlugin       unsafe.Pointer
        stopPlugin        unsafe.Pointer
        reloadPlugin      unsafe.Pointer
        finiPlugin        unsafe.Pointer
        command           unsafe.Pointer
        freePlugin        unsafe.Pointer
        eventListenerLoop unsafe.Pointer
        eventResult       unsafe.Pointer
}

type dynLoadLibManager struct {
	dynLoadLibs map[string]*dynLoadLib
	rwMutex	*sync.RWMutex
}

func (d *dynLoadLibManager) newDynLoadLib(pluginName string, dlHandle *dlHandle,
     getBuildVersion unsafe.Pointer, getName unsafe.Pointer, newPlugin unsafe.Pointer,
     initPlugin unsafe.Pointer, startPlugin unsafe.Pointer, stopPlugin unsafe.Pointer,
     reloadPlugin unsafe.Pointer, finiPlugin unsafe.Pointer, command unsafe.Pointer,
     freePlugin unsafe.Pointer, eventListenerLoop unsafe.Pointer,
     eventResult unsafe.Pointer) *dynLoadLib {
	return &dynLoadLib {
		pluginName        : pluginName,
		dlHandle          : dlHandle,
		getBuildVersion   : getBuildVersion,
		getName           : getName,
		newPlugin         : newPlugin,
		initPlugin        : initPlugin,
		startPlugin       : startPlugin,
		stopPlugin        : stopPlugin,
		reloadPlugin      : reloadPlugin,
		finiPlugin        : finiPlugin,
		command           : command,
		freePlugin        : freePlugin,
		eventListenerLoop : eventListenerLoop,
		eventResult       : eventResult,
	}
}

func (d *dynLoadLibManager) get(pluginName string) (*dynLoadLib, bool) {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
	dlLib, ok := d.dynLoadLibs[pluginName]
	return dlLib, ok 
}

func (d *dynLoadLibManager) setIfAbsent(pluginName string, dlLib *dynLoadLib) bool {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	 _, ok := d.dynLoadLibs[pluginName]
	if !ok {
		 d.dynLoadLibs[pluginName] = dlLib
	}
	return !ok
}

func (d *dynLoadLibManager) delete(pluginName string) {
        d.rwMutex.Lock()
        defer d.rwMutex.Unlock()
	delete(d.dynLoadLibs, pluginName)
}

func (d *dynLoadLibManager) getPluginNames() []string {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
        pluginNames := make([]string, 0, 0)
        for pluginName, _ := range d.dynLoadLibs {
                pluginNames = append(pluginNames, pluginName)
        }
        return pluginNames
}

func (d *dynLoadLibManager) foreachPluginNames(cbfunc func(pluginName string) bool) {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
        for pluginName, _ := range d.dynLoadLibs {
		if !cbfunc(pluginName) {
			break
		}
        }
}

func (d *dynLoadLibManager) foreachDynLoadLibs(cbfunc func(dlLib *dynLoadLib) bool) {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
        for _, dlLib := range d.dynLoadLibs {
		if !cbfunc(dlLib) {
			break
		}
        }
}

func (d *dynLoadLibManager) clear() {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.dynLoadLibs = make(map[string]*dynLoadLib)
}

func newDynLoadLibManager() *dynLoadLibManager {
	return &dynLoadLibManager {
		dynLoadLibs : make(map[string]*dynLoadLib),
		rwMutex : new(sync.RWMutex),
	}
}

type PluginHandle struct {
        pluginName           string
        instanceId           uint64
        dlLib                *dynLoadLib
        coder                *common.Coder
        newResultFunc        func() interface{}
        newCommandResultFunc func() interface{}
        newEventParamFunc    func() interface{}
        eventHandlerMgr      *eventHandlerManager
        smallBufferMgr       *bufferManager
        bigBufferMgr         *bufferManager
	requestIdGen         *idGenerator
	eventIdGen           *idGenerator
}

type pluginHandleManager struct {
	pluginHandles map[uint64]*PluginHandle
	rwMutex	*sync.RWMutex
}

func (p *pluginHandleManager) newPluginHandle(instanceId uint64,
    dlLib *dynLoadLib, pluginName string, newResultFunc func() interface{},
    newCommandResultFunc func() interface{}, newEventParamFunc func() interface{}) *PluginHandle {
        return &PluginHandle {
                instanceId: instanceId,
                dlLib: dlLib,
                pluginName: pluginName,
                coder: common.NewCoder(),
                newResultFunc : newResultFunc,
                newCommandResultFunc : newCommandResultFunc,
                newEventParamFunc : newEventParamFunc,
                eventHandlerMgr : newEventHandlerManager(),
                smallBufferMgr : newBufferManager(smallBufferSize),
                bigBufferMgr : newBufferManager(bigBufferSize),
                requestIdGen : newIdGenerator(),
                eventIdGen : newIdGenerator(),
        }
}

func (p *pluginHandleManager) get(instanceId uint64) (*PluginHandle, bool) {
        p.rwMutex.RLock()
        defer p.rwMutex.RUnlock()
        pluginHandle, ok := p.pluginHandles[instanceId]
	return pluginHandle, ok
}

func (p *pluginHandleManager) set(instanceId uint64, pluginHandle *PluginHandle) {
        p.rwMutex.Lock()
        defer p.rwMutex.Unlock()
        p.pluginHandles[instanceId] = pluginHandle
}

func (p *pluginHandleManager) delete(instanceId uint64) {
        p.rwMutex.Lock()
        defer p.rwMutex.Unlock()
	delete(p.pluginHandles, instanceId)
}

func (p *pluginHandleManager) foreachInstPlugHandle(cbfunc func(instanceId uint64, pluginHandle *PluginHandle) bool) {
        p.rwMutex.RLock()
        defer p.rwMutex.RUnlock()
        for instanceId, pluginHandle := range p.pluginHandles {
		if !cbfunc(instanceId, pluginHandle) {
			break
		}
	}
}

func (p *pluginHandleManager) clear() {
        p.rwMutex.Lock()
        defer p.rwMutex.Unlock()
	p.pluginHandles = make(map[uint64]*PluginHandle)
}

func newPluginHandleManager() *pluginHandleManager {
	return &pluginHandleManager{
		pluginHandles : make(map[uint64]*PluginHandle),
		rwMutex : new(sync.RWMutex),
	}
}

type eventHandlerManager struct {
	eventHandlers map[string]func(pluginHandle *PluginHandle,
	    eventName string, eventParam interface{}, err error) (interface{}, error)
	rwMutex	*sync.RWMutex
}

func (e *eventHandlerManager) set(eventName string,
     eventHandler func(pluginHandle *PluginHandle, eventName string, eventParam interface{}, err error) (interface{}, error)) {
        e.rwMutex.Lock()
        defer e.rwMutex.Unlock()
	e.eventHandlers[eventName] = eventHandler
}

func (e *eventHandlerManager) get(eventName string) (func(pluginHandle *PluginHandle,
     eventName string, eventParam interface{}, err error) (interface{}, error), bool) {
        e.rwMutex.Lock()
        defer e.rwMutex.Unlock()
	eventHandler, ok := e.eventHandlers[eventName]
	return eventHandler, ok
}

func newEventHandlerManager() *eventHandlerManager {
	return &eventHandlerManager{
		eventHandlers : make(map[string]func(pluginHandle *PluginHandle,
		    eventName string, eventParam interface{}, err error) (interface{}, error)),
		rwMutex : new(sync.RWMutex),
	}
}
