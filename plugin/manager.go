package plugin

import (
	"sync"
	"github.com/pkg/errors"
	"github.com/potix/plugger/coder"
	"github.com/potix/plugger/generator"
)


type eventInfo struct {
	eventId          uint64
	eventName        *string
	eventParamBuffer *[]byte
	err              error
}

type eventManager struct {
	eventMap map[uint64]*eventInfo
}

func (em *eventManager) addEvent(eventId uint64, eventName *string, eventParamBuffer *[]byte, err error) {
	eventInfo := &eventInfo {
		eventId          : eventId,
		eventName        : eventName,
		eventParamBuffer : eventParamBuffer,
		err              : err,
	}
	em.eventMap[eventId] = eventInfo
}

func (em *eventManager) getEvent(eventId uint64) (*eventInfo, error) {
	eventInfo, ok := em.eventMap[eventId]
	if !ok {
		return nil, errors.Errorf("not found event (id = %v)", eventId)
	}
	return eventInfo, nil
}

func (em *eventManager) deleteEvent(eventId uint64) {
	delete(em.eventMap, eventId)
}

type resultInfo struct {
	resultId     uint64
	resultBuffer *[]byte
	err          error
}

type resultManager struct {
	resultMap map[uint64]*resultInfo
	mutex     *sync.Mutex
}

func (rm *resultManager) addResult(resultId uint64, result *[]byte, err error) {
	rm.mutex.Lock()
        defer rm.mutex.Unlock()
	resultInfo := &resultInfo {
		resultId     : resultId,
		resultBuffer : result,
		err          : err,
	}
	rm.resultMap[resultId] = resultInfo
}

func (rm *resultManager) getResult(resultId uint64) (*resultInfo, error) {
	rm.mutex.Lock()
        defer rm.mutex.Unlock()
	result, ok := rm.resultMap[resultId]
	if !ok {
		return nil, errors.Errorf("not found result (id = %v)", resultId)
	}
	return result, nil
}

func (rm *resultManager) deleteResult(resultId uint64) {
	rm.mutex.Lock()
        defer rm.mutex.Unlock()
	delete(rm.resultMap, resultId)
}

type instanceInfo struct {
	instanceId    uint64
	plugin        Plugin
	resultManager *resultManager
	eventManager  *eventManager
	eventChan     chan *event
}

type instManager struct {
	instMap         map[uint64]*instanceInfo
	instMapByPlugin map[Plugin]*instanceInfo
	rwMutex         *sync.RWMutex
}

func (im *instManager) addInstance(instanceId uint64, plugin Plugin) {
	im.rwMutex.Lock()
        defer im.rwMutex.Unlock()
	instInfo := &instanceInfo {
		instanceId      : instanceId,
		plugin          : plugin,
		resultManager   : new(resultManager),
		eventManager    : new(eventManager),
		eventChan       : make(chan *event),
	}
	im.instMap[instanceId] = instInfo
	im.instMapByPlugin[plugin] = instInfo
}

func (im *instManager) getInstance(instanceId uint64) (*instanceInfo, error) {
	im.rwMutex.RLock()
        defer im.rwMutex.RUnlock()
	instInfo, ok := im.instMap[instanceId]
	if !ok {
		return nil, errors.Errorf("not found instance (id = %v)", instanceId)
	}
	return instInfo, nil
}

func (im *instManager) getInstanceByPlugin(plugin Plugin) (*instanceInfo, error) {
	im.rwMutex.RLock()
        defer im.rwMutex.RUnlock()
	instInfo, ok := im.instMapByPlugin[plugin]
	if !ok {
		return nil, errors.Errorf("not found instance (plugin = %v)", plugin)
	}
	return instInfo, nil
}

func (im *instManager) deleteInstance(instanceId uint64) {
	im.rwMutex.Lock()
        defer im.rwMutex.Unlock()
	instInfo, ok := im.instMap[instanceId]
	if !ok {
		return 
	}
	close(instInfo.eventChan)
	delete(im.instMap, instanceId)
	delete(im.instMapByPlugin, instInfo.plugin)
}

type pluginManager struct {
	pluginName          string
	newPluginFunc       func() Plugin
	newConfigFunc       func() interface{}
	newCommandParamFunc func() interface{}
	// TODO
	//   event
	instMgr             *instManager
	coder               *coder.Coder
	idGen               *generator.IdGenerator
}

func (pm *pluginManager) eventEmit(plugin Plugin, event *event) error {
	instInfo, err := pm.instMgr.getInstanceByPlugin(plugin)
	if err != nil {
		return err
	}
	instInfo.eventChan <- event
	return nil
}

var pluginMgr *pluginManager

func init() {
	pluginMgr = &pluginManager {
		instMgr : &instManager {
			instMap         : make(map[uint64]*instanceInfo),
			instMapByPlugin : make(map[Plugin]*instanceInfo),
			rwMutex         : new(sync.RWMutex),
		},
		coder   : coder.NewCoder(),
		idGen   : generator.NewIdGenerator(),
	}
}
