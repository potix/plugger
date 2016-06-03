package plugin

import "C"

import (
	"fmt"
	"bytes"
	"reflect"
	"unsafe"
	"github.com/pkg/errors"
	"github.com/potix/plugger/common"
)

//export GetBuildVersion
func GetBuildVersion() C.ulonglong {
	return common.BuildVersion
}

//export GetName
func GetName(namePtr *C.char, nameLen *C.int) C.int {
	// recall
	if len(pluginMgr.pluginName) > int(*nameLen) {
		// lack of name buffer
		// need recall
		return C.int(common.RVNeedGlowPluginNameBuffer)
	}
	copyBytesToPtrAndLen([]byte(pluginMgr.pluginName), namePtr, nameLen)
	return C.int(common.RVSuccess)
}

//export NewPlugin
func NewPlugin(instanceId C.ulonglong) C.int {
	if pluginMgr.newPluginFunc == nil {
		return C.int(common.RVNotImplemented)
	}
	plugin := pluginMgr.newPluginFunc()
	pluginMgr.instMgr.setInstance(uint64(instanceId), plugin)
	return C.int(common.RVSuccess)
}

//export InitPlugin
func InitPlugin(instanceId C.ulonglong, recall C.int, requestId C.ulonglong,
     resultPtr *C.char, resultLen *C.int,
    errorPtr *C.char, errorLen *C.int,
     argPtr *C.char, argLen C.int) C.int {
	if pluginMgr.newConfigFunc == nil {
		// not implemented
		return C.int(common.RVNotImplemented)
	}
	return withArgBaseFunc(instanceId, recall, requestId,
	    resultPtr, resultLen,
	    errorPtr, errorLen,
	    argPtr, argLen,
	    pluginMgr.newConfigFunc(), "Init")
}

//export StartPlugin
func StartPlugin(instanceId C.ulonglong, recall C.int, requestId C.ulonglong, 
     resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int) C.int {
	return noArgBaseFunc(instanceId, recall, requestId,
            resultPtr, resultLen,
	    errorPtr, errorLen, "Start")
}

//export StopPlugin
func StopPlugin(instanceId C.ulonglong, recall C.int, requestId C.ulonglong, 
     resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int) C.int {
	return noArgBaseFunc(instanceId, recall, requestId,
            resultPtr, resultLen,
	    errorPtr, errorLen, "Stop")
}

//export ReloadPlugin
func ReloadPlugin(instanceId C.ulonglong, recall C.int, requestId C.ulonglong,
     resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int,
     argPtr *C.char, argLen C.int) C.int {
	if pluginMgr.newConfigFunc == nil {
		// not implemented
		return C.int(common.RVNotImplemented)
	}
	return withArgBaseFunc(instanceId, recall, requestId,
	    resultPtr, resultLen,
            errorPtr, errorLen,
	    argPtr, argLen,
	    pluginMgr.newConfigFunc(), "Reload")
}

//export FiniPlugin
func FiniPlugin(instanceId C.ulonglong, recall C.int, requestId C.ulonglong, 
     resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int) C.int {
	return noArgBaseFunc(instanceId, recall, requestId,
            resultPtr, resultLen,
	    errorPtr, errorLen, "Fini")
}

//export Command
func Command(instanceId C.ulonglong, recall C.int, requestId C.ulonglong,
     resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int,
     argPtr *C.char, argLen C.int) C.int {
	if pluginMgr.newCommandParamFunc == nil {
		// not implemented
		return C.int(common.RVNotImplemented)
	}
	return withArgBaseFunc(instanceId, recall, requestId,
	    resultPtr, resultLen,
            errorPtr, errorLen, argPtr,
	    argLen, pluginMgr.newCommandParamFunc(), "Command")
}

//export FreePlugin
func FreePlugin(instanceId C.ulonglong) C.int {
	pluginMgr.instMgr.deleteInstance(uint64(instanceId))
	return C.int(common.RVSuccess)
}

//export EventListenerLoop
func EventListenerLoop(instanceId C.ulonglong, recall C.int, eventId C.ulonglong, 
     eventNamePtr *C.char, eventNameLen *C.int,
     eventParamPtr *C.char, eventParamLen *C.int,
     errorPtr *C.char, errorLen *C.int) C.int {
	return eventListenerLoopFunc(instanceId, recall, eventId,
            eventNamePtr, eventNameLen,
	    eventParamPtr, eventParamLen,
	    errorPtr, errorLen)
}

//export EventResult
func EventResult(instanceId C.ulonglong,
     eventId C.ulonglong, eventHandling C.int,
     eventResultPtr *C.char, eventResultLen C.int,
     argErrorPtr *C.char, argErrorLen C.int) C.int {
	return eventResultFunc(instanceId, 
            eventId, eventHandling,
            eventResultPtr, eventResultLen,
	    argErrorPtr, argErrorLen,
            pluginMgr.newEventResultFunc)
}

func noArgBaseFunc(instanceId C.ulonglong, recall C.int, requestId C.ulonglong,
     resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int, method string) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	if recall != 0 {
		// recall
		resultInfo, err := instInfo.resultMgr.get(uint64(requestId))
		if err != nil {
			// not reached
			panic(fmt.Sprintf("not found result of instance (requestId = %v)", requestId))
		}
		if len(*resultInfo.resultBuffer) != 0 && len(*resultInfo.resultBuffer) > int(*resultLen) {
			// lack of result buffer
			// need recall
			return C.int(common.RVNeedGlowResultBuffer)
		}
		if resultInfo.err != nil && len(resultInfo.err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			return C.int(common.RVNeedGlowErrorBuffer)
		}
		if len(*resultInfo.resultBuffer) != 0 {
			copyBytesToPtrAndLen(*resultInfo.resultBuffer, resultPtr, resultLen)
		} else {
			*resultLen = 0
		}
		if resultInfo.err != nil {
			copyBytesToPtrAndLen([]byte(resultInfo.err.Error()), errorPtr, errorLen)
		} else {
			*errorLen = 0
		}
		instInfo.resultMgr.delete(uint64(requestId))
		return C.int(common.RVSuccess)
	}
	resultBuffer := make([]byte, 0, 0)
	fn := getPluginMethodbyName(instInfo.plugin, method).(func() (interface{}, error))
        result, err0 := fn()
	if result != nil {
       		resultBytesBuffer := bytes.NewBuffer(make([]byte, 0, bigBufferSize))
		err := pluginMgr.coder.Encode(result, resultBytesBuffer)
		if err != nil {
			if err0 != nil {
				err0 = errors.Wrap(err0, err.Error())
			} else {
				err0 = err
			}
		} else {
			resultBuffer = resultBytesBuffer.Bytes()
		}
	}
	if  len(resultBuffer) != 0 && len(resultBuffer) > int(*resultLen) {
		// lack of result buffer
		// need recall
		instInfo.resultMgr.add(uint64(requestId),
		    instInfo.resultMgr.newResultInfo(uint64(requestId), &resultBuffer, err0))
		return C.int(common.RVNeedGlowResultBuffer)
	}
	if err0 != nil && len(err0.Error()) > int(*errorLen) {
		// lack of error buffer
		// need recall
		instInfo.resultMgr.add(uint64(requestId),
		    instInfo.resultMgr.newResultInfo(uint64(requestId), &resultBuffer, err0))
		return C.int(common.RVNeedGlowErrorBuffer)
	}
	if len(resultBuffer) != 0 {
		copyBytesToPtrAndLen(resultBuffer, resultPtr, resultLen)
	} else {
		*resultLen = 0;
	}
	if err0 != nil {
		copyBytesToPtrAndLen([]byte(err0.Error()), errorPtr, errorLen)
	} else {
		*errorLen = 0
	}
        return C.int(common.RVSuccess)
}

func withArgBaseFunc(instanceId C.ulonglong, recall C.int, requestId C.ulonglong,
     resultPtr *C.char, resultLen *C.int, 
     errorPtr *C.char, errorLen *C.int,
     argPtr *C.char, argLen C.int, value interface{}, method string) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	if recall != 0 {
		// recall
		resultInfo, err := instInfo.resultMgr.get(uint64(requestId))
		if err != nil {
			// not reached
			panic(fmt.Sprintf("not found result of instance (requestId = %v)", requestId))
		}
		if len(*resultInfo.resultBuffer) != 0 && len(*resultInfo.resultBuffer) > int(*resultLen) {
			// lack of result buffer
			// need recall
			return C.int(common.RVNeedGlowResultBuffer)
		}
		if resultInfo.err != nil && len(resultInfo.err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			return C.int(common.RVNeedGlowErrorBuffer)
		}
		if len(*resultInfo.resultBuffer) != 0 {
			copyBytesToPtrAndLen(*resultInfo.resultBuffer, resultPtr, resultLen)
		} else {
			*resultLen = 0
		}
		if resultInfo.err != nil {
			copyBytesToPtrAndLen([]byte(resultInfo.err.Error()), errorPtr, errorLen)
		} else {
			*errorLen = 0
		}
		instInfo.resultMgr.delete(uint64(requestId))
		return C.int(common.RVSuccess)
	}
	var resultBuffer []byte = make([]byte, 0, 0)
	param := ptrAndLenToBytes(argPtr, argLen)
	if err := pluginMgr.coder.Decode(bytes.NewBuffer(param), value); err != nil {
		if err != nil && len(err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			instInfo.resultMgr.add(uint64(requestId),
			    instInfo.resultMgr.newResultInfo(uint64(requestId), &resultBuffer, err))
			return C.int(common.RVNeedGlowErrorBuffer)
		}
		*resultLen = 0
		copyBytesToPtrAndLen([]byte(err.Error()), errorPtr, errorLen)
		return C.int(common.RVSuccess)
	} 
	fn := getPluginMethodbyName(instInfo.plugin, method).(func(interface{})(interface{}, error))
        result, err0 := fn(value)
	if result != nil {
       		resultBytesBuffer := bytes.NewBuffer(make([]byte, 0, bigBufferSize))
		err := pluginMgr.coder.Encode(result, resultBytesBuffer)
		if err != nil {
			if err0 != nil {
				err0 = errors.Wrap(err0, err.Error())
			} else {
				err0 = err
			}
		} else {
			resultBuffer = resultBytesBuffer.Bytes()
		}
	}
	if  len(resultBuffer) != 0 && len(resultBuffer) > int(*resultLen) {
		// lack of result buffer
		// need recall
		instInfo.resultMgr.add(uint64(requestId),
		    instInfo.resultMgr.newResultInfo(uint64(requestId), &resultBuffer, err0))
		return C.int(common.RVNeedGlowResultBuffer)
	}
	if err0 != nil && len(err0.Error()) > int(*errorLen) {
		// lack of error buffer
		// need recall
		instInfo.resultMgr.add(uint64(requestId),
		    instInfo.resultMgr.newResultInfo(uint64(requestId), &resultBuffer, err0))
		return C.int(common.RVNeedGlowErrorBuffer)
	}
	if len(resultBuffer) != 0 {
		copyBytesToPtrAndLen(resultBuffer, resultPtr, resultLen)
	} else {
		*resultLen = 0
	}
	if err0 != nil {
		copyBytesToPtrAndLen([]byte(err0.Error()), errorPtr, errorLen)
	} else {
		*errorLen = 0
	}
        return C.int(common.RVSuccess)
}

func eventListenerLoopFunc(instanceId C.ulonglong, recall C.int, eventId C.ulonglong,
      eventNamePtr *C.char, eventNameLen *C.int,
      eventParamPtr *C.char, eventParamLen *C.int,
      errorPtr *C.char, errorLen *C.int,) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	if recall != 0 {
		// recall
		eventInfo, err := instInfo.eventMgr.get(uint64(eventId))
		if err != nil {
			// not reached
			panic(fmt.Sprintf("not found result of instance (eventId = %v)", eventId))
		}
		if len(*eventInfo.eventName) != 0 && len(*eventInfo.eventName) > int(*eventNameLen) {
			// lack of event name buffer
			// need recall
			return C.int(common.RVNeedGlowEventNameBuffer)
		}
		if len(*eventInfo.eventParamBuffer) != 0 && len(*eventInfo.eventParamBuffer) > int(*eventParamLen) {
			// lack of result buffer
			// need recall
			return C.int(common.RVNeedGlowEventParamBuffer)
		}
		if eventInfo.err != nil && len(eventInfo.err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			return C.int(common.RVNeedGlowErrorBuffer)
		}
		if len(*eventInfo.eventName) != 0 {
			copyBytesToPtrAndLen([]byte(*eventInfo.eventName), eventNamePtr, eventNameLen)
		} else {
			*eventNameLen = 0
		}
		if len(*eventInfo.eventParamBuffer) != 0 {
			copyBytesToPtrAndLen(*eventInfo.eventParamBuffer, eventParamPtr, eventParamLen)
		} else {
			*eventParamLen = 0
		}
		if eventInfo.err != nil {
			copyBytesToPtrAndLen([]byte(eventInfo.err.Error()), errorPtr, errorLen)
		} else {
			*errorLen = 0
		}
		return C.int(common.RVSuccess)
	}
	event, ok := <-instInfo.eventReqChan
	if !ok {
		// free instance
		// channel close
		return C.int(common.RVFinishEventListenerLoop)
	}
	eventParamBuffer := make([]byte, 0, 0)
	if event.eventParam != nil {
       		eventParamBytesBuffer := bytes.NewBuffer(make([]byte, 0, bigBufferSize))
		err = pluginMgr.coder.Encode(event.eventParam, eventParamBytesBuffer)
		if err == nil {
			eventParamBuffer = eventParamBytesBuffer.Bytes()
		}
	}
	if  len(event.eventName) != 0 && len(event.eventName) > int(*eventNameLen) {
		// lack of event name buffer
		// need recall
		instInfo.eventMgr.set(uint64(eventId), instInfo.eventMgr.newEventInfo(uint64(eventId),
		    &event.eventName, &eventParamBuffer, err, event.eventResChan))
		return C.int(common.RVNeedGlowEventNameBuffer)
	}
	if  len(eventParamBuffer) != 0 && len(eventParamBuffer) > int(*eventParamLen) {
		// lack of event param buffer
		// need recall
		instInfo.eventMgr.set(uint64(eventId), instInfo.eventMgr.newEventInfo(uint64(eventId),
		    &event.eventName, &eventParamBuffer, err, event.eventResChan))
		return C.int(common.RVNeedGlowEventParamBuffer)
	}
	if err != nil && len(err.Error()) > int(*errorLen) {
		// lack of error buffer
		// need recall
		instInfo.eventMgr.set(uint64(eventId), instInfo.eventMgr.newEventInfo(uint64(eventId),
		    &event.eventName, &eventParamBuffer, err, event.eventResChan))
		return C.int(common.RVNeedGlowErrorBuffer)
	}
	if len(event.eventName) != 0 {
		copyBytesToPtrAndLen([]byte(event.eventName), eventNamePtr, eventNameLen)
	} else {
		*eventNameLen = 0
	}
	if len(eventParamBuffer) != 0 {
		copyBytesToPtrAndLen(eventParamBuffer, eventParamPtr, eventParamLen)
	} else {
		*eventParamLen = 0
	}
	if err != nil {
		copyBytesToPtrAndLen([]byte(err.Error()), errorPtr, errorLen)
	} else {
		*errorLen = 0
	}
	instInfo.eventMgr.set(uint64(eventId),
	    instInfo.eventMgr.newEventInfo(uint64(eventId), nil, nil, nil, event.eventResChan))
        return C.int(common.RVSuccess)
}

func eventResultFunc(instanceId C.ulonglong,
      eventId C.ulonglong, eventHandling C.int,
      eventResultPtr *C.char, eventResultLen C.int,
      argErrorPtr *C.char, argErrorLen C.int,
      eventResultFunc func() interface{}) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	eventInfo, err := instInfo.eventMgr.get(uint64(eventId))
	if err != nil {
		// not reached
		panic(fmt.Sprintf("not found result of instance (eventId = %v)", eventId))
	}
	if (eventResultFunc == nil) {
		eventInfo.eventResChan <- instInfo.eventMgr.newEventResponse(
		    nil, errors.New("not implemented"), common.EVNotImplemented)
		instInfo.eventMgr.delete(uint64(eventId))
		return C.int(common.RVNotImplemented)
	}
	eventResult := eventResultFunc()
	if (common.EventHandling(eventHandling) == common.EVNoHandling) {
		eventInfo.eventResChan <- instInfo.eventMgr.newEventResponse(
		    eventResult, errors.New("no handling"), common.EVNoHandling)
		instInfo.eventMgr.delete(uint64(eventId))
		return C.int(common.RVSuccess)
	} else if (common.EventHandling(eventHandling) == common.EVEncodeError) {
		eventInfo.eventResChan <- instInfo.eventMgr.newEventResponse(
		    eventResult,  errors.New("encode errors"), common.EVEncodeError)
		instInfo.eventMgr.delete(uint64(eventId))
		return C.int(common.RVSuccess)
	}
	eventResultBuffer := ptrAndLenToBytes(eventResultPtr, eventResultLen)
	if err := pluginMgr.coder.Decode(bytes.NewBuffer(eventResultBuffer), eventResult); err != nil {
		eventInfo.eventResChan <- instInfo.eventMgr.newEventResponse(
			eventResult, err, common.EVDecodeError)
		instInfo.eventMgr.delete(uint64(eventId))
		return C.int(common.RVSuccess)
	} 
	var argError error
	if argErrorLen != 0 {
		argErrorBuffer := ptrAndLenToBytes(argErrorPtr, argErrorLen)
		argError = errors.New(string(argErrorBuffer))
	}
	eventInfo.eventResChan <- instInfo.eventMgr.newEventResponse(
	    eventResult, argError, common.EventHandling(eventHandling))
	instInfo.eventMgr.delete(uint64(eventId))
        return C.int(common.RVSuccess)
}

func copyBytesToPtrAndLen(src []byte, ptr *C.char, ptrLen *C.int) {
	var pos uintptr = 0
	for _, b := range src {
		*(*uint8)(unsafe.Pointer((uintptr(unsafe.Pointer(ptr)) + pos))) = uint8(b)
		pos += 1
	}
	*ptrLen = C.int(len(src))
}

func getPluginMethodbyName(plugin Plugin, method string) interface{} {
	return reflect.ValueOf(plugin).MethodByName(method).Interface()
}

func ptrAndLenToBytes(ptr *C.char, ptrLen C.int) []byte {
        return C.GoBytes(unsafe.Pointer(ptr), ptrLen)
}
