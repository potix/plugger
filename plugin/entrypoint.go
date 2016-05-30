package plugin

import "C"

import (
	"fmt"
	"bytes"
	"reflect"
	"unsafe"
	"github.com/pkg/errors"
        "github.com/potix/plugger/definer"
)

//export IsMatchName
func IsMatchName(requestedPluginName *C.char) C.int {
        if C.GoString(requestedPluginName) == pluginMgr.pluginName {
                return 1
        } else {
                return 0
        }
}

//export NewPlugin
func NewPlugin(instanceId C.ulonglong) {
	plugin := pluginMgr.newPluginFunc()
	pluginMgr.instMgr.addInstance(uint64(instanceId), plugin)
}

//export InitPlugin
func InitPlugin(instanceId C.ulonglong, resultId *C.ulonglong, resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int, argPtr *C.char, argLen C.int) C.int {
	return withArgBaseFunc(instanceId, resultId, resultPtr, resultLen,
	    errorPtr, errorLen, argPtr, argLen, pluginMgr.newConfigFunc(), "Init")
}

//export StartPlugin
func StartPlugin(instanceId C.ulonglong, resultId *C.ulonglong, 
     resultPtr *C.char, resultLen *C.int, errorPtr *C.char, errorLen *C.int) C.int {
	return noArgBaseFunc(instanceId, resultId,
            resultPtr, resultLen, errorPtr, errorLen, "Start")
}

//export StopPlugin
func StopPlugin(instanceId C.ulonglong, resultId *C.ulonglong, 
     resultPtr *C.char, resultLen *C.int, errorPtr *C.char, errorLen *C.int) C.int {
	return noArgBaseFunc(instanceId, resultId,
            resultPtr, resultLen, errorPtr, errorLen, "Stop")
}

//export ReloadPlugin
func ReloadPlugin(instanceId C.ulonglong, resultId *C.ulonglong, resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int, argPtr *C.char, argLen C.int) C.int {
	return withArgBaseFunc(instanceId, resultId, resultPtr, resultLen,
            errorPtr, errorLen, argPtr, argLen, pluginMgr.newConfigFunc(), "Reload")
}

//export FiniPlugin
func FiniPlugin(instanceId C.ulonglong, resultId *C.ulonglong, 
     resultPtr *C.char, resultLen *C.int, errorPtr *C.char, errorLen *C.int) C.int {
	return noArgBaseFunc(instanceId, resultId,
            resultPtr, resultLen, errorPtr, errorLen, "Fini")
}

//export Command
func Command(instanceId C.ulonglong, resultId *C.ulonglong, resultPtr *C.char, resultLen *C.int,
     errorPtr *C.char, errorLen *C.int, argPtr *C.char, argLen C.int) C.int {
	return withArgBaseFunc(instanceId, resultId, resultPtr, resultLen,
            errorPtr, errorLen, argPtr, argLen, pluginMgr.newCommandParamFunc(), "Command")
}

//export FreePlugin
func FreePlugin(instanceId C.ulonglong) {
	pluginMgr.instMgr.deleteInstance(uint64(instanceId))
}

//export EventListenerLoop
func EventListenerLoop(instanceId C.ulonglong, eventId *C.ulonglong, 
     eventNamePtr *C.char, eventNameLen *C.int, eventParamPtr *C.char, eventParamLen *C.int,
     errorPtr *C.char, errorLen *C.int) C.int {
	return eventListenerLoopFunc(instanceId, eventId,
            eventNamePtr, eventNameLen, eventParamPtr, eventParamLen,
	    errorPtr, errorLen)
}

func noArgBaseFunc(instanceId C.ulonglong, resultId *C.ulonglong,
      resultPtr *C.char, resultLen *C.int, errorPtr *C.char, errorLen *C.int, method string) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	if *resultId != 0 {
		// recall
		resultInfo, err := instInfo.resultManager.getResult(uint64(*resultId))
		if err != nil {
			// not reached
			panic(fmt.Sprintf("not found result of instance (resultId = %v)", resultId))
		}
		if len(*resultInfo.resultBuffer) != 0 && len(*resultInfo.resultBuffer) > int(*resultLen) {
			// lack of result buffer
			// need recall
			return 1
		}
		if resultInfo.err != nil && len(resultInfo.err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			return 2
		}
		if len(*resultInfo.resultBuffer) != 0 {
			copyBytesToPtrAndLen(*resultInfo.resultBuffer, resultPtr, resultLen)
		} else {
			*resultLen = 0;
		}
		if resultInfo.err != nil {
			copyBytesToPtrAndLen([]byte(resultInfo.err.Error()), errorPtr, errorLen)
		} else {
			*errorLen = 0
		}
		instInfo.resultManager.deleteResult(uint64(*resultId))
		return 0
	}
	resultBuffer := make([]byte, 0, 0)
	fn := getPluginMethodbyName(instInfo.plugin, method).(func() (interface{}, error))
        result, err0 := fn()
	if result != nil {
       		resultBytesBuffer := bytes.NewBuffer(make([]byte, 0, definer.ResultBufferSize))
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
		newResultId := pluginMgr.idGen.Get()
		instInfo.resultManager.addResult(newResultId, &resultBuffer, err0)
		*resultId = C.ulonglong(newResultId)
		return 1
	}
	if err0 != nil && len(err0.Error()) > int(*errorLen) {
		// lack of error buffer
		// need recall
		newResultId := pluginMgr.idGen.Get()
		instInfo.resultManager.addResult(newResultId, &resultBuffer, err0)
		*resultId = C.ulonglong(newResultId)
		return 2
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
        return 0
}

func withArgBaseFunc(instanceId C.ulonglong, resultId *C.ulonglong, resultPtr *C.char, resultLen *C.int, 
     errorPtr *C.char, errorLen *C.int, argPtr *C.char, argLen C.int, value interface{}, method string) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	if *resultId != 0 {
		// recall
		resultInfo, err := instInfo.resultManager.getResult(uint64(*resultId))
		if err != nil {
			// not reached
			panic(fmt.Sprintf("not found result of instance (resultId = %v)", resultId))
		}
		if len(*resultInfo.resultBuffer) != 0 && len(*resultInfo.resultBuffer) > int(*resultLen) {
			// lack of result buffer
			// need recall
			return 1
		}
		if resultInfo.err != nil && len(resultInfo.err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			return 2
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
		instInfo.resultManager.deleteResult(uint64(*resultId))
		return 0
	}
	var resultBuffer []byte = make([]byte, 0, 0)
	param := ptrAndLenToBytes(argPtr, argLen)
	if err := pluginMgr.coder.Decode(bytes.NewBuffer(param), value); err != nil {
		if err != nil && len(err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			newResultId := pluginMgr.idGen.Get()
			instInfo.resultManager.addResult(newResultId, &resultBuffer, err)
			return 2
		}
		*resultLen = 0
		copyBytesToPtrAndLen([]byte(err.Error()), errorPtr, errorLen)
		return 0
	} 
	fn := getPluginMethodbyName(instInfo.plugin, method).(func(interface{})(interface{}, error))
        result, err0 := fn(value)
	if result != nil {
       		resultBytesBuffer := bytes.NewBuffer(make([]byte, 0, definer.ResultBufferSize))
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
		newResultId := pluginMgr.idGen.Get()
		instInfo.resultManager.addResult(newResultId, &resultBuffer, err0)
		*resultId = C.ulonglong(newResultId)
		return 1
	}
	if err0 != nil && len(err0.Error()) > int(*errorLen) {
		// lack of error buffer
		// need recall
		newResultId := pluginMgr.idGen.Get()
		instInfo.resultManager.addResult(newResultId, &resultBuffer, err0)
		*resultId = C.ulonglong(newResultId)
		return 2
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
        return 0
}

func eventListenerLoopFunc(instanceId C.ulonglong, eventId *C.ulonglong,
      eventNamePtr *C.char, eventNameLen *C.int, eventParamPtr *C.char, eventParamLen *C.int,
      errorPtr *C.char, errorLen *C.int,) C.int {
        instInfo, err := pluginMgr.instMgr.getInstance(uint64(instanceId))
        if err != nil {
		// not reached
		panic(fmt.Sprintf("not found instance of plugin (instanceId = %v)", instanceId))
        }
	if *eventId != 0 {
		// recall
		eventInfo, err := instInfo.eventManager.getEvent(uint64(*eventId))
		if err != nil {
			// not reached
			panic(fmt.Sprintf("not found result of instance (eventId = %v)", eventId))
		}
		if len(*eventInfo.eventName) != 0 && len(*eventInfo.eventName) > int(*eventNameLen) {
			// lack of event name buffer
			// need recall
			return 3
		}
		if len(*eventInfo.eventParamBuffer) != 0 && len(*eventInfo.eventParamBuffer) > int(*eventParamLen) {
			// lack of result buffer
			// need recall
			return 4
		}
		if eventInfo.err != nil && len(eventInfo.err.Error()) > int(*errorLen) {
			// lack of error buffer
			// need recall
			return 2
		}
		if len(*eventInfo.eventName) != 0 {
			copyBytesToPtrAndLen([]byte(*eventInfo.eventName), eventNamePtr, eventNameLen)
		} else {
			*eventNameLen = 0
		}
		if len(*eventInfo.eventParamBuffer) != 0 {
			copyBytesToPtrAndLen(*eventInfo.eventParamBuffer, eventParamPtr, eventParamLen)
		} else {
			*eventParamLen = 0;
		}
		if eventInfo.err != nil {
			copyBytesToPtrAndLen([]byte(eventInfo.err.Error()), errorPtr, errorLen)
		} else {
			*errorLen = 0
		}
		instInfo.eventManager.deleteEvent(uint64(*eventId))
		return 0
	}
	event, ok := <-instInfo.eventChan
	if !ok {
		// free instance
		// channel close
		return 5
	}
	eventParamBuffer := make([]byte, 0, 0)
	if event.eventParam != nil {
       		eventParamBytesBuffer := bytes.NewBuffer(make([]byte, 0, definer.EventParamBufferSize))
		err = pluginMgr.coder.Encode(event.eventParam, eventParamBytesBuffer)
		if err == nil {
			eventParamBuffer = eventParamBytesBuffer.Bytes()
		}
	}
	if  len(event.eventName) != 0 && len(event.eventName) > int(*eventNameLen) {
		// lack of event name buffer
		// need recall
		newEventId := pluginMgr.idGen.Get()
		instInfo.eventManager.addEvent(newEventId, &event.eventName, &eventParamBuffer, err)
		*eventId = C.ulonglong(newEventId)
		return 3
	}
	if  len(eventParamBuffer) != 0 && len(eventParamBuffer) > int(*eventParamLen) {
		// lack of event param buffer
		// need recall
		newEventId := pluginMgr.idGen.Get()
		instInfo.eventManager.addEvent(newEventId, &event.eventName, &eventParamBuffer, err)
		*eventId = C.ulonglong(newEventId)
		return 4
	}
	if err != nil && len(err.Error()) > int(*errorLen) {
		// lack of error buffer
		// need recall
		newEventId := pluginMgr.idGen.Get()
		instInfo.eventManager.addEvent(newEventId, &event.eventName, &eventParamBuffer, err)
		*eventId = C.ulonglong(newEventId)
		return 2
	}
	if len(event.eventName) != 0 {
		copyBytesToPtrAndLen([]byte(event.eventName), eventNamePtr, eventNameLen)
	} else {
		*eventNameLen = 0
	}
	if len(eventParamBuffer) != 0 {
		copyBytesToPtrAndLen(eventParamBuffer, eventParamPtr, eventParamLen)
	} else {
		*eventParamLen = 0;
	}
	if err != nil {
		copyBytesToPtrAndLen([]byte(err.Error()), errorPtr, errorLen)
	} else {
		*errorLen = 0
	}
        return 0
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
