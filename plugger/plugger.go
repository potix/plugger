package plugger 

// #include <stdio.h>
// #include <stdlib.h>
//
// unsigned long long bridge_get_build_version(void *fnptr)
// {
//	unsigned long long (*fn)() = fnptr;
//	return fn();
// }
//
// int bridge_get_name(void *fnptr, char *name, int *name_len)
// {
//	int (*fn)(char *, int *) = fnptr;
//	return fn(name, name_len);
// }
//
// bridge_new_free(void *fnptr, unsigned long long id)
// {
//	int (*fn)(unsigned long long) = fnptr;
//	return fn(id);
// }
//
// int bridge_no_argument_func(
//     void *fnptr,
//     unsigned long long instance_id, int recall,
//     unsigned long long request_id,
//     char *result_ptr, int *result_len,
//     char *error_ptr, int *error_len)
// {
//	int (*fn)(unsigned long long, int recall, unsigned long long,
//          char *, int *, char *, int *) = fnptr;
//	return fn(instance_id, recall, request_id,
//          result_ptr, result_len,
//	    error_ptr, error_len);
// }
//
// int bridge_with_argument_func(
//     void *fnptr,
//     unsigned long long instance_id, int recall,
//     unsigned long long request_id,
//     char *result_ptr, int *result_len,
//     char *error_ptr, int *error_len,
//     char *arg_ptr, int arg_len)
// {
//	int (*fn)(unsigned long long, int recall, unsigned long long,
//          char *, int *, char *, int *, char *, int) = fnptr;
//	return fn(instance_id, recall, request_id,
//	    result_ptr, result_len,
//          error_ptr, error_len,
//	    arg_ptr, arg_len);
// }
//
// int bridge_elloop_func(
//     void *fnptr,
//     unsigned long long instance_id, int recall,
//     unsigned long long event_id,
//     char *event_name_ptr, int *event_name_len,
//     char *event_param_ptr, int *event_param_len,
//     char *error_ptr, int *error_len)
// {
//	int (*fn)(unsigned long long, int recall, unsigned long long,
//          char *, int *, char *, int *, char *, int *) = fnptr;
//	return fn(instance_id, recall, event_id,
//          event_name_ptr, event_name_len,
//          event_param_ptr, event_param_len,
//	    error_ptr, error_len);
// }
//
// int bridge_event_result_func(
//     void *fnptr,
//     unsigned long long instance_id,
//     unsigned long long event_id,
//     int event_handling,
//     char *event_result_ptr, int event_result_len,
//     char *arg_error_ptr, int arg_error_len)
// {
//	int (*fn)(unsigned long long,
//          unsigned long long, int,
//          char *, int, char *, int) = fnptr;
//	return fn(instance_id,
//	    event_id, event_handling,
//          event_result_ptr, event_result_len,
//	    arg_error_ptr, arg_error_len);
// }
import "C"

import (
	"fmt"
	"os"
	"bytes"
	"unsafe"
	"reflect"
	"runtime"
	"github.com/pkg/errors"
        "github.com/potix/plugger/common"
)

func (ph *PluginHandle)GetPluginName() string {
	return ph.pluginName
}

func (ph *PluginHandle)Init(config interface{}) (result interface{}, err error) {
	if ph.newResultFunc == nil {
		return nil, errors.New("not implemented method")
	}
	return ph.callEpWithArg(ph.dlLib.initPlugin, config, ph.newResultFunc())
}

func (ph *PluginHandle)Start() (result interface{}, err  error) {
	if ph.newResultFunc == nil {
		return nil, errors.New("not implemented method")
	}
	return ph.callEpNoArg(ph.dlLib.startPlugin, ph.newResultFunc())
}

func (ph *PluginHandle)Stop() (result interface{}, err error) {
	if ph.newResultFunc == nil {
		return nil, errors.New("not implemented method")
	}
	return ph.callEpNoArg(ph.dlLib.stopPlugin, ph.newResultFunc())
}

func (ph *PluginHandle)Reload(config interface{}) (result interface{}, err error) {
	if ph.newResultFunc == nil {
		return nil, errors.New("not implemented method")
	}
	return ph.callEpWithArg(ph.dlLib.reloadPlugin, config, ph.newResultFunc())
}

func (ph *PluginHandle)Fini() (result interface{}, err error) {
	if ph.newResultFunc == nil {
		return nil, errors.New("not implemented method")
	}
	return ph.callEpNoArg(ph.dlLib.finiPlugin, ph.newResultFunc())
}

func (ph *PluginHandle)Command(commandParam interface{}) (commandResult interface{}, err error) {
	if ph.newCommandResultFunc == nil {
		return nil, errors.New("not implemented method")
	}
	return ph.callEpWithArg(ph.dlLib.command, commandParam, ph.newCommandResultFunc())
}

func (ph *PluginHandle)EventOn(eventName string,
     eventHandler func(pluginHandle *PluginHandle, eventName string,
     eventParam interface{}, err error) (eventResult interface{}, handleErr error)) {
	ph.eventHandlerMgr.set(eventName, eventHandler)
}

func (ph *PluginHandle) eventListenerLoop() {
	if ph.newEventParamFunc == nil {
		return
	}
	for {
		eventId, eventName, eventParam, err, finish := ph.callEpELLoop(ph.dlLib.eventListenerLoop, ph.newEventParamFunc())
		if finish {
			break
		}
		eventHandler, ok := ph.eventHandlerMgr.get(eventName)
		if !ok {
			ph.callEpEventResult(ph.dlLib.eventResult, eventId, nil, nil, common.EVNoHandling)
		}
		eventResult, err := eventHandler(ph, eventName, eventParam, err)
		ph.callEpEventResult(ph.dlLib.eventResult, eventId, eventResult, err, common.EVHandling)
	}
}

func (ph *PluginHandle)callEpNoArg(callPtr unsafe.Pointer, result interface{}) (interface{}, error) {
	// call entry point
	var recall int
	var requestId uint64 = ph.requestIdGen.Get() 
	resultBuffer := ph.bigBufferMgr.getBuffer()
	errorBuffer := ph.smallBufferMgr.getBuffer()
	resultBufferPtr, resultBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(resultBuffer)
	errorBufferPtr, errorBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(errorBuffer)
	for {
		retval := C.bridge_no_argument_func(callPtr,
		    C.ulonglong(ph.instanceId),
		    C.int(recall),
		    C.ulonglong(requestId),
		    resultBufferPtr, resultBufferLenPtr,
		    errorBufferPtr, errorBufferLenPtr)
		if common.ReturnValue(retval) == common.RVSuccess {
			// ok
			break
		} else if common.ReturnValue(retval) == common.RVNeedGlowResultBuffer {
			// more result buffer
			// glow buffer and recall
			resultBuffer = ph.bigBufferMgr.glowBuffer(resultBuffer)
			resultBufferPtr, resultBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(resultBuffer)
			recall = 1
		} else if common.ReturnValue(retval) == common.RVNeedGlowErrorBuffer {
			// more error buffer
			// glow buffer and recall
			errorBuffer = ph.smallBufferMgr.glowBuffer(errorBuffer)
			errorBufferPtr, errorBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(errorBuffer)
			recall = 1
		}
	}
	var err error
	resultBufferLen := getIntFromCintPtr(resultBufferLenPtr)
	errorBufferLen := getIntFromCintPtr(errorBufferLenPtr)
	// rebuild error
	if errorBufferLen != 0 {
		err = errors.New(string(errorBuffer[0:errorBufferLen]))
	}
	// rebuild result
	if resultBufferLen != 0 {
		decodeErr := ph.coder.Decode(bytes.NewBuffer(resultBuffer[0:resultBufferLen]), result)
		if decodeErr != nil {
			if err == nil {
				 err = decodeErr
			} else {
				 err = errors.Wrap(err, decodeErr.Error())
			}
		}
	}
	return result, err
}

func (ph *PluginHandle)callEpWithArg(callPtr unsafe.Pointer, param interface{}, result interface{}) (interface{}, error) {
	// call entry point
	argBytesBuffer := bytes.NewBuffer(make([]byte, 0, bigBufferSize))
	if param != nil {
		if err := ph.coder.Encode(param, argBytesBuffer); err != nil {
			return result, errors.Wrap(err, "could't encode parameter")
		}
	}
	argBuffer := argBytesBuffer.Bytes()
	argBufferPtr, _ := getCucharPtrAndCintPtrFromByteSlice(argBuffer)
	argBufferLen := len(argBuffer)
	
	var recall int
	var requestId uint64 = ph.requestIdGen.Get() 
	resultBuffer := ph.bigBufferMgr.getBuffer()
	errorBuffer := ph.smallBufferMgr.getBuffer()
	resultBufferPtr, resultBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(resultBuffer)
	errorBufferPtr, errorBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(errorBuffer)
	for {
		retval := C.bridge_with_argument_func(callPtr,
		    C.ulonglong(ph.instanceId),
		    C.int(recall),
		    C.ulonglong(requestId),
		    resultBufferPtr, resultBufferLenPtr,
		    errorBufferPtr, errorBufferLenPtr,
		    argBufferPtr, C.int(argBufferLen))
		if common.ReturnValue(retval) == common.RVSuccess {
			// ok
			break
		} else if common.ReturnValue(retval) == common.RVNeedGlowResultBuffer {
			// more result buffer
			// glow buffer and recall
			resultBuffer = ph.bigBufferMgr.glowBuffer(resultBuffer)
			resultBufferPtr, resultBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(resultBuffer)
			recall = 1
		} else if common.ReturnValue(retval) == common.RVNeedGlowErrorBuffer {
			// more error buffer
			// glow buffer and recall
			errorBuffer = ph.smallBufferMgr.glowBuffer(errorBuffer)
			errorBufferPtr, errorBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(errorBuffer)
			recall = 1
		} else if common.ReturnValue(retval) == common.RVNotImplemented {
			// not implemented
			return nil, errors.New("not implemented method")
		}
	}
	var err error
	resultBufferLen := getIntFromCintPtr(resultBufferLenPtr)
	errorBufferLen := getIntFromCintPtr(errorBufferLenPtr)
	// rebuild error
	if errorBufferLen != 0 {
		err = errors.New(string(errorBuffer[0:errorBufferLen]))
	}
	// rebuild result
	if resultBufferLen != 0 {
		decodeErr := ph.coder.Decode(bytes.NewBuffer(resultBuffer[0:resultBufferLen]), result)
		if decodeErr != nil {
			if err == nil {
				err = decodeErr
			} else {
				err = errors.Wrap(err, decodeErr.Error())
			}
		}
	}
	return result, err
}

func (ph *PluginHandle)callEpELLoop(callPtr unsafe.Pointer, eventParam interface{}) (uint64, string, interface{}, error, bool) {
	// call entry point
	var recall int
	var eventId uint64 = ph.eventIdGen.Get()
	eventParamBuffer := ph.bigBufferMgr.getBuffer()
	eventNameBuffer := ph.smallBufferMgr.getBuffer()
	errorBuffer := ph.smallBufferMgr.getBuffer()
	eventParamBufferPtr, eventParamBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(eventParamBuffer)
	eventNameBufferPtr, eventNameBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(eventNameBuffer)
	errorBufferPtr, errorBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(eventNameBuffer)
	for {
		retval := C.bridge_elloop_func(callPtr,
		    C.ulonglong(ph.instanceId),
		    C.int(recall),
		    C.ulonglong(eventId),
		    eventNameBufferPtr, eventNameBufferLenPtr,
		    eventParamBufferPtr, eventParamBufferLenPtr,
		    errorBufferPtr, errorBufferLenPtr)
		if common.ReturnValue(retval) == common.RVSuccess {
			// ok
			break
		} else if common.ReturnValue(retval) == common.RVNeedGlowEventNameBuffer {
			// more event name buffer
			// glow buffer and recall
			eventNameBuffer = ph.bigBufferMgr.glowBuffer(eventNameBuffer)
			eventNameBufferPtr, eventNameBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(eventNameBuffer)
			recall = 1
		} else if common.ReturnValue(retval) == common.RVNeedGlowEventParamBuffer {
			// more event param buffer
			// glow buffer and recall
			eventParamBuffer = ph.bigBufferMgr.glowBuffer(eventParamBuffer)
			eventParamBufferPtr, eventParamBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(eventParamBuffer)
			recall = 1
		} else if common.ReturnValue(retval) == common.RVNeedGlowErrorBuffer {
			// more error buffer
			// glow buffer and recall
			errorBuffer = ph.bigBufferMgr.glowBuffer(errorBuffer)
			errorBufferPtr, errorBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(errorBuffer)
			recall = 1
		} else if common.ReturnValue(retval) == common.RVFinishEventListenerLoop {
			// finish
			return eventId, "", eventParam, nil, true
		}
	}
	var err error
	eventParamBufferLen := getIntFromCintPtr(eventParamBufferLenPtr)
	eventNameBufferLen := getIntFromCintPtr(eventNameBufferLenPtr)
	errorBufferLen := getIntFromCintPtr(errorBufferLenPtr)
	// rebuild error
	if errorBufferLen != 0 {
		err = errors.New(string(errorBuffer[0:errorBufferLen]))
	}
	// rebuild event name
	var eventName string
	if eventNameBufferLen != 0 {
		eventName = string(eventNameBuffer[0:eventNameBufferLen])
	}
	// rebuild event param
	if eventParamBufferLen != 0 {
		decodeErr := ph.coder.Decode(bytes.NewBuffer(eventParamBuffer[0:eventParamBufferLen]), eventParam)
		if decodeErr != nil {
			if err == nil {
				err = decodeErr
			} else {
				err = errors.Wrap(err, decodeErr.Error())
			}
		}
	}
	return eventId, eventName, eventParam, err, false
}

func (ph *PluginHandle)callEpEventResult(callPtr unsafe.Pointer, eventId uint64,
     eventResult interface{}, err error, eventHandling common.EventHandling) {
	// call entry point
	eventResultBytesBuffer := bytes.NewBuffer(make([]byte, 0, bigBufferSize))
	if eventResult != nil {
		if err := ph.coder.Encode(eventResult, eventResultBytesBuffer); err != nil {
			eventHandling = common.EVEncodeError
		}
	}
	eventResultBuffer := eventResultBytesBuffer.Bytes()
	eventResultBufferPtr, _ := getCucharPtrAndCintPtrFromByteSlice(eventResultBuffer)
	eventResultBufferLen := len(eventResultBuffer)
	argErrorBytesBuffer := bytes.NewBuffer(make([]byte, 0, smallBufferSize))
	if err != nil {
		if err := ph.coder.Encode(err.Error(), argErrorBytesBuffer); err != nil {
			eventHandling = common.EVEncodeError
		}
	}
	argErrorBuffer := argErrorBytesBuffer.Bytes()
	argErrorBufferPtr, _ := getCucharPtrAndCintPtrFromByteSlice(argErrorBuffer)
	argErrorBufferLen := len(argErrorBuffer)
	retval := C.bridge_event_result_func(callPtr,
	    C.ulonglong(ph.instanceId),
	    C.ulonglong(eventId),
	    C.int(eventHandling),
	    eventResultBufferPtr, C.int(eventResultBufferLen),
	    argErrorBufferPtr, C.int(argErrorBufferLen))
	if common.ReturnValue(retval) == common.RVSuccess {
		// ok
		return
	} else if common.ReturnValue(retval) == common.RVNotImplemented {
		// not implemented
		return
	} else {
		// not reached
		panic("unexpected return value")
		return
	}
}

type Plugger struct {
	dynLoadLibMgr   *dynLoadLibManager
	pluginHandleMgr *pluginHandleManager
	instanceIdGen   *idGenerator
	smallBufferMgr  *bufferManager
	versionSafe     bool
}

func (p *Plugger) SetVersionSafe(versionSafe bool) {
	p.versionSafe = versionSafe
}

func (p *Plugger) Load(pluginDirPath string) (err error, warn error) {
	var warn error
	ext := ".so"
	if runtime.GOOS == "windows" {
		return errors.New("no support platform")
	} else if runtime.GOOS == "darwin" {
		ext = ".dylib"
	}
	pluginFiles := make([]string, 0, 0)
	if err := gatherPluginFiles(&pluginFiles, pluginDirPath, ext); err != nil {
		return err, nil
	}
	for _, pf := range pluginFiles {
		dlHandle, err := dlOpen(pf, Now|Local)
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		getBuildVersion, err := dlHandle.dlSymbol("GetBuildVersion")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		getName, err := dlHandle.dlSymbol("GetName")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		newPlugin, err := dlHandle.dlSymbol("NewPlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		initPlugin, err := dlHandle.dlSymbol("InitPlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		startPlugin, err := dlHandle.dlSymbol("StartPlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		stopPlugin, err := dlHandle.dlSymbol("StopPlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		reloadPlugin, err := dlHandle.dlSymbol("ReloadPlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		finiPlugin, err := dlHandle.dlSymbol("FiniPlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		command, err := dlHandle.dlSymbol("Command")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		freePlugin, err := dlHandle.dlSymbol("FreePlugin")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		eventListenerLoop, err := dlHandle.dlSymbol("EventListenerLoop")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		eventResult, err := dlHandle.dlSymbol("EventResult")
		if err != nil {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: %v", pf, err.Error()))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: %v", pf, err.Error()))
			}
			dlHandle.dlClose()
			continue
		}
		pluginName := p.callEpGetName(getName)
		if pluginName == "" {
			if warn != nil {
				warn = errors.New(fmt.Sprintf("%v: plugin name is empty", pf))
			} else {
				errors.Wrap(warn, fmt.Sprintf("%v: plugin name is empty", pf))
			}
			dlHandle.dlClose()
			continue
		}
		pluginBuildVersion := p.callEpGetBuildVersion(getBuildVersion)
		if common.BuildVersion != pluginBuildVersion {
			if p.versionSafe {
				if warn != nil {
					warn = fmt.Sprintf("build version is mismatch (plugger %v, plugin %v),skip plugin",
					    common.BuildVersion, pluginBuildVersion)
				} else {
					errors.Wrap(warn,
					    fmt.Sprintf("build version is mismatch (plugger %v, plugin %v),skip plugin",
					    common.BuildVersion, pluginBuildVersion))
				}
				dlHandle.dlClose()
				continue
			}
		}
	 	if !p.dynLoadLibMgr.setIfAbsent(pluginName, p.dynLoadLibMgr.newDynLoadLib(pluginName, dlHandle,
		    getBuildVersion, getName, newPlugin, initPlugin, startPlugin, stopPlugin, reloadPlugin,
		    finiPlugin, command, freePlugin, eventListenerLoop, eventResult)) {
			// already exist plugin
			dlHandle.dlClose()
			continue
		}
	}
	return nil, warn
}

func (p *Plugger) GetBuildVersion() uint64 {
	return common.BuildVersion
}

func (p *Plugger) GetPluginNames() (pluginNames []string) {
	return p.dynLoadLibMgr.getPluginNames()
}

func (p *Plugger) ExistsPluginNames(candidatePluginNames []string) (pluginNames []string) {
	exists := make([]string, 0, 0)
	for _, candidatePluginName := range candidatePluginNames {
		p.dynLoadLibMgr.foreachPluginNames(func(pluginName string) bool {
			if candidatePluginName == pluginName {
				exists = append(exists, pluginName)
				return false
			}
			return true
		})
	}
	return exists
}

func (p *Plugger) NewPlugin(pluginName string, newResultFunc func() interface{},
     newCommandResultFunc func() interface{}, newEventParamFunc func() interface{}) (pluginHandle *PluginHandle, err error) {
	dlLib, ok := p.dynLoadLibMgr.get(pluginName)
	if !ok {
		return nil, errors.Errorf("not found plugin name (%v)", pluginName)
	}
	instanceId := p.instanceIdGen.Get()
	if C.bridge_new_free(dlLib.newPlugin, C.ulonglong(instanceId)) == -1 {
		return nil, errors.Errorf("not implemented")
	}
	ph := p.pluginHandleMgr.newPluginHandle(instanceId,
	    dlLib, pluginName, newResultFunc, newCommandResultFunc, newEventParamFunc)
	p.pluginHandleMgr.set(instanceId, ph)
	for i := 0; i < runtime.NumCPU(); i++ {
		go ph.eventListenerLoop()
	}
	return ph, nil
}

func (p *Plugger) Unload(pluginName string) (err error) {
	dlLib, ok := p.dynLoadLibMgr.get(pluginName)
	if !ok {
		return errors.Errorf("not found plugin name (%v)", pluginName)
	}
	instIds := make([]uint64, 0, 0)
	p.pluginHandleMgr.foreachInstPlugHandle(func(instanceId uint64, pluginHandle *PluginHandle) bool {
		if pluginHandle.pluginName == pluginName {
			C.bridge_new_free(dlLib.freePlugin, C.ulonglong(pluginHandle.instanceId))
			instIds = append(instIds, instanceId)
		}
		return true
	})
	for _, instId := range instIds {
		p.pluginHandleMgr.delete(instId)
	}
	dlLib.dlHandle.dlClose()
	p.dynLoadLibMgr.delete(pluginName)
	return nil
}

func (p *Plugger) FreePlugin(pluginHandle *PluginHandle) (err error) {
	dlLib, ok := p.dynLoadLibMgr.get(pluginHandle.pluginName)
	if !ok {
		return errors.Errorf("not found plugin name (%v)", pluginHandle.pluginName)
	}
	if _, ok := p.pluginHandleMgr.get(pluginHandle.instanceId); !ok {
		return errors.Errorf("not found plugin handle (%v)", pluginHandle.instanceId)
	}
	C.bridge_new_free(dlLib.freePlugin, C.ulonglong(pluginHandle.instanceId))
	p.pluginHandleMgr.delete(pluginHandle.instanceId)
	return nil
}

func (p *Plugger) Free() {
	p.pluginHandleMgr.foreachInstPlugHandle(func(instanceId uint64, pluginHandle *PluginHandle) bool {
		C.bridge_new_free(pluginHandle.dlLib.freePlugin, C.ulonglong(instanceId))
		return true
	})
	p.pluginHandleMgr.clear()
	p.dynLoadLibMgr.foreachDynLoadLibs(func(dlLib *dynLoadLib) bool {
		dlLib.dlHandle.dlClose()
		return true
	})
	p.dynLoadLibMgr.clear()
}

func (p *Plugger)callEpGetBuildVersion(callPtr unsafe.Pointer) uint64 {
	return uint64(C.bridge_get_build_version(callPtr))
}

func (p *Plugger)callEpGetName(callPtr unsafe.Pointer) string {
	// call entry point
	nameBuffer := p.smallBufferMgr.getBuffer()
	nameBufferPtr, nameBufferLenPtr := getCucharPtrAndCintPtrFromByteSlice(nameBuffer)
	for {
		retval := C.bridge_get_name(callPtr, nameBufferPtr, nameBufferLenPtr)
		if common.ReturnValue(retval) == common.RVSuccess {
			// ok
			break
		} else if common.ReturnValue(retval) == common.RVNeedGlowPluginNameBuffer {
			// more name buffer
			// glow buffer and recall
			nameBuffer = p.smallBufferMgr.glowBuffer(nameBuffer)
			nameBufferPtr, nameBufferLenPtr = getCucharPtrAndCintPtrFromByteSlice(nameBuffer)
		}
	}
	// rebuild name
	var name string
	var nameBufferLen = getIntFromCintPtr(nameBufferLenPtr)
	if nameBufferLen != 0 {
		name = string(nameBuffer[0:nameBufferLen])
	} 
	return name
}

func NewPlugger() *Plugger {
	p := new(Plugger)
	p.dynLoadLibMgr = newDynLoadLibManager()
	p.pluginHandleMgr = newPluginHandleManager()
	p.instanceIdGen = newIdGenerator()
	p.smallBufferMgr = newBufferManager(smallBufferSize)
	return p
}

func getCucharPtrAndCintPtrFromByteSlice(b []byte) (*C.char, *C.int) {
	var l = cap(b)
	return (*C.char)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)), (*C.int)(unsafe.Pointer(&l))
}

func getIntFromCintPtr(v *C.int) int {
	return *(*int)(unsafe.Pointer(v))
}
