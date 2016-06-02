package plugin

import (
	"sync"
	"github.com/pkg/errors"
	"github.com/potix/plugger/common"
)

const (
	bigBufferSize = 2048
)

type eventRequest struct {
        eventName string
        eventParam interface{}
	eventResChan chan *eventResponse
}

type eventResponse struct {
        eventResult interface{}
        err error
	eventHandling common.EventHandling
}

type eventInfo struct {
	eventId          uint64
	eventName        *string
	eventParamBuffer *[]byte
	err              error
	eventResChan     chan *eventResponse
	argErr           error
}

type eventManager struct {
	eventMap map[uint64]*eventInfo
	rwMutex   *sync.RWMutex
}

func (em *eventManager) newEventResponse(
    eventResult interface{}, err error, eventHandling common.EventHandling) *eventResponse {
	return &eventResponse {
		eventResult: eventResult,
		err: err,
		eventHandling: eventHandling,
	}
}

func (em *eventManager) newEventInfo(eventId uint64, eventName *string,
     eventParamBuffer *[]byte, err error, eventResChan chan *eventResponse, argErr error) *eventInfo {
	return &eventInfo {
		eventId          : eventId,
		eventName        : eventName,
		eventParamBuffer : eventParamBuffer,
		err              : err,
		eventResChan     : eventResChan,
		argErr           : argErr,
	}
}

func (em *eventManager) set(eventId uint64, eventInfo *eventInfo) {
	em.rwMutex.Lock()
        defer em.rwMutex.Unlock()
	em.eventMap[eventId] = eventInfo
}

func (em *eventManager) get(eventId uint64) (*eventInfo, error) {
	em.rwMutex.RLock()
        defer em.rwMutex.RUnlock()
	eventInfo, ok := em.eventMap[eventId]
	if !ok {
		return nil, errors.Errorf("not found event (id = %v)", eventId)
	}
	return eventInfo, nil
}

func (em *eventManager) delete(eventId uint64) {
	em.rwMutex.Lock()
        defer em.rwMutex.Unlock()
	delete(em.eventMap, eventId)
}

type resultInfo struct {
	resultId     uint64
	resultBuffer *[]byte
	err          error
}

type resultManager struct {
	resultMap map[uint64]*resultInfo
	rwMutex     *sync.RWMutex
}

func (rm *resultManager) newResultInfo(resultId uint64, result *[]byte, err error) *resultInfo {
	return &resultInfo {
		resultId     : resultId,
		resultBuffer : result,
		err          : err,
	}
}

func (rm *resultManager) add(resultId uint64, resultInfo *resultInfo) {
	rm.rwMutex.Lock()
        defer rm.rwMutex.Unlock()
	rm.resultMap[resultId] = resultInfo
}

func (rm *resultManager) get(resultId uint64) (*resultInfo, error) {
	rm.rwMutex.RLock()
        defer rm.rwMutex.RUnlock()
	result, ok := rm.resultMap[resultId]
	if !ok {
		return nil, errors.Errorf("not found result (id = %v)", resultId)
	}
	return result, nil
}

func (rm *resultManager) delete(resultId uint64) {
	rm.rwMutex.Lock()
        defer rm.rwMutex.Unlock()
	delete(rm.resultMap, resultId)
}

type instanceInfo struct {
	instanceId    uint64
	plugin        Plugin
	resultMgr     *resultManager
	eventMgr      *eventManager
	eventReqChan  chan *eventRequest
}

type instManager struct {
	instMap         map[uint64]*instanceInfo
	instMapByPlugin map[Plugin]*instanceInfo
	rwMutex         *sync.RWMutex
}

func (im *instManager) setInstance(instanceId uint64, plugin Plugin) {
	im.rwMutex.Lock()
        defer im.rwMutex.Unlock()
	instInfo := &instanceInfo {
		instanceId   : instanceId,
		plugin       : plugin,
		resultMgr    : &resultManager {
			resultMap : make(map[uint64]*resultInfo),
			rwMutex   : new(sync.RWMutex),
		},
		eventMgr     : &eventManager {
			eventMap : make(map[uint64]*eventInfo),
			rwMutex  : new(sync.RWMutex),
		},
		eventReqChan : make(chan *eventRequest),
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
	close(instInfo.eventReqChan)
	delete(im.instMap, instanceId)
	delete(im.instMapByPlugin, instInfo.plugin)
}

type pluginManager struct {
	pluginName          string
	newPluginFunc       func() Plugin
	newConfigFunc       func() interface{}
	newCommandParamFunc func() interface{}
	newEventResultFunc  func() interface{}
	instMgr             *instManager
	coder               *common.Coder
}

func (pm *pluginManager) eventEmit(plugin Plugin, eventRequest *eventRequest) (interface{}, error) {
	instInfo, err := pm.instMgr.getInstanceByPlugin(plugin)
	if err != nil {
		return nil, err
	}
	instInfo.eventReqChan <- eventRequest
	eventResponse := <- eventRequest.eventResChan
	return eventResponse.eventResult, eventResponse.err
}

var pluginMgr *pluginManager

func init() {
	pluginMgr = &pluginManager {
		instMgr : &instManager {
			instMap         : make(map[uint64]*instanceInfo),
			instMapByPlugin : make(map[Plugin]*instanceInfo),
			rwMutex         : new(sync.RWMutex),
		},
		coder   : common.NewCoder(),
	}
}
