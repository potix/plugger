package plugger 

// #include <stdio.h>
// #include <stdlib.h>
//
// int bridge_ismatch(void *fnptr, char *name)
// {
//	int (*fn)(char *) = fnptr;
//	return fn(name);
// }
//
// bridge_id_only(void *fnptr, unsigned long long id)
// {
//	long long (*fn)(unsigned long long) = fnptr;
//	fn(id);
// }
//
// int bridge_no_argument_func(
//     void *fnptr,
//     unsigned long long instance_id, unsigned long long *result_id,
//     char *result_ptr, int *result_len,
//     char *error_ptr, int *error_len)
// {
//	int (*fn)(unsigned long long, unsigned long long *,
//          char *, int *, char *, int *) = fnptr;
//	return fn(instance_id, result_id,
//          result_ptr, result_len, error_ptr, error_len);
// }
//
// int bridge_with_argument_func(
//     void *fnptr,
//     unsigned long long instance_id, unsigned long long *result_id,
//     char *result_ptr, int *result_len,
//     char *error_ptr, int *error_len,
//     char *arg_ptr, int arg_len)
// {
//	int (*fn)(unsigned long long, unsigned long long *,
//          char *, int *, char *, int *, char *, int) = fnptr;
//	return fn(instance_id, result_id, result_ptr, result_len,
//          error_ptr, error_len, arg_ptr, arg_len);
// }
//
// int bridge_elloop_func(
//     void *fnptr,
//     unsigned long long instance_id, unsigned long long *event_id,
//     char *event_name_ptr, int *event_name_len,
//     char *event_param_ptr, int *event_param_len,
//     char *error_ptr, int *error_len)
// {
//	int (*fn)(unsigned long long, unsigned long long *,
//          char *, int *, char *, int *, char *, int *) = fnptr;
//	return fn(instance_id, event_id,
//          event_name_ptr, event_name_len,
//          event_param_ptr, event_param_len,  error_ptr, error_len);
// }
import "C"

import (
	"os"
	"bytes"
	"unsafe"
	"reflect"
	"github.com/pkg/errors"
        "github.com/potix/plugger/coder"
        "github.com/potix/plugger/definer"
        "github.com/potix/plugger/generator"
)

type dynLoadLib struct {
	pluginName        string
	dlHandle          *dlHandle
	isMatchName       unsafe.Pointer
	newPlugin         unsafe.Pointer
	initPlugin        unsafe.Pointer
	startPlugin       unsafe.Pointer
	stopPlugin        unsafe.Pointer
	reloadPlugin      unsafe.Pointer
	finiPlugin        unsafe.Pointer
	command           unsafe.Pointer
	freePlugin        unsafe.Pointer
	eventListenerLoop unsafe.Pointer
}

type PluginHandle struct {
	pluginName        string
	instanceId        uint64
	dlLib             *dynLoadLib
	resultBuffer      []byte
	errorBuffer       []byte
	eventNameBuffer   []byte
	eventParamBuffer  []byte
	coder             *coder.Coder
	newResultFunc     func() interface{}
	newEventParamFunc func() interface{}
	eventHandlers     map[string]func(pluginHandle *PluginHandle, eventName string, eventParam interface{}, err error) 
}

func (ph *PluginHandle)GetPluginName() string {
	return ph.pluginName
}

func (ph *PluginHandle)Init(config interface{}) (interface{}, error) {
	return ph.callEpWithArg(ph.dlLib.initPlugin, config, ph.newResultFunc())
}

func (ph *PluginHandle)Start() (interface{},  error) {
	return ph.callEpNoArg(ph.dlLib.startPlugin, ph.newResultFunc())
}

func (ph *PluginHandle)Stop() (interface{}, error) {
	return ph.callEpNoArg(ph.dlLib.startPlugin, ph.newResultFunc())
}

func (ph *PluginHandle)Reload(config interface{}) (interface{}, error) {
	return ph.callEpWithArg(ph.dlLib.reloadPlugin, config, ph.newResultFunc())
}

func (ph *PluginHandle)Fini() (interface{}, error) {
	return ph.callEpNoArg(ph.dlLib.finiPlugin, ph.newResultFunc() )
}

func (ph *PluginHandle)Command(commandParam interface{}) (interface{}, error) {
	return ph.callEpWithArg(ph.dlLib.command, commandParam, ph.newResultFunc())
}

func (ph *PluginHandle)EventOn(eventName string, eventHandler func(pluginHandle *PluginHandle, eventName string, eventParam interface{}, err error)) {
	ph.eventHandlers[eventName] = eventHandler
}

func (ph *PluginHandle) eventListenerLoop() {
	for {
		eventName, eventParam, err, finish := ph.callEpELLoop(ph.dlLib.eventListenerLoop, ph.newEventParamFunc())
		if finish {
			break
		}
		eventHandler, ok := ph.eventHandlers[eventName]
		if !ok {
			// no registered event handler
			// discard event
			continue
		}
		eventHandler(ph, eventName, eventParam, err)
	}
}

func (ph *PluginHandle)callEpNoArg(callPtr unsafe.Pointer, result interface{}) (interface{}, error) {
	// call entry point
	var resultId uint64 = 0
	var resultBufferLen int = cap(ph.resultBuffer) 
	var errorBufferLen int = cap(ph.errorBuffer)
	resultPtr := getCucharPtrFromByteSlicePtr(&ph.resultBuffer)
	errorPtr := getCucharPtrFromByteSlicePtr(&ph.errorBuffer)
	for {
		code := C.bridge_no_argument_func(callPtr, C.ulonglong(ph.instanceId), getCulonglongPtrFromUint64Ptr(&resultId),
		    resultPtr, getCintPtrFromIntPtr(&resultBufferLen),
		    errorPtr, getCintPtrFromIntPtr(&errorBufferLen))
		if code == 0 {
			// ok
			break
		} else if code == 1 {
			// more result buffer
			// glow buffer and recall
			ph.resultBuffer = make([]byte, 0, cap(ph.resultBuffer) * 2)
			resultPtr = getCucharPtrFromByteSlicePtr(&ph.resultBuffer)
			resultBufferLen = cap(ph.resultBuffer)
		} else if code == 2 {
			// more error buffer
			// glow buffer and recall
			ph.errorBuffer = make([]byte, 0, cap(ph.errorBuffer) * 2)
			errorPtr = getCucharPtrFromByteSlicePtr(&ph.errorBuffer)
			errorBufferLen = cap(ph.errorBuffer)
		}
	}
	// rebuild error
	var err error = nil
	if errorBufferLen != 0 {
		err = errors.New(string(ph.errorBuffer[0:errorBufferLen]))
	}
	// rebuild result
	if resultBufferLen != 0 {
		decodeErr := ph.coder.Decode(bytes.NewBuffer(ph.resultBuffer[0:resultBufferLen]), result)
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
	argBytesBuffer := bytes.NewBuffer(make([]byte, 0, definer.ParamBufferSize))
	err := ph.coder.Encode(param, argBytesBuffer)
	if err != nil {
		return result, errors.Wrap(err, "could't encode parameter")
	}
	argBuffer := argBytesBuffer.Bytes()
	var resultId uint64 = 0
	var resultBufferLen int = cap(ph.resultBuffer) 
	var errorBufferLen int = cap(ph.errorBuffer)
	var argBufferLen int = len(argBuffer)
	resultPtr := getCucharPtrFromByteSlicePtr(&ph.resultBuffer)
	errorPtr := getCucharPtrFromByteSlicePtr(&ph.errorBuffer)
	for {
		code := C.bridge_with_argument_func(callPtr, C.ulonglong(ph.instanceId), getCulonglongPtrFromUint64Ptr(&resultId),
		    resultPtr, getCintPtrFromIntPtr(&resultBufferLen),
		    errorPtr, getCintPtrFromIntPtr(&errorBufferLen),
		    getCucharPtrFromByteSlicePtr(&argBuffer), C.int(argBufferLen))
		if code == 0 {
			// ok
			break
		} else if code == 1 {
			// more result buffer
			// glow buffer and recall
			ph.resultBuffer = make([]byte, 0, cap(ph.resultBuffer) * 2)
			resultPtr = getCucharPtrFromByteSlicePtr(&ph.resultBuffer)
			resultBufferLen = cap(ph.resultBuffer)
		} else if code == 2 {
			// more error buffer
			// glow buffer and recall
			ph.errorBuffer = make([]byte, 0, cap(ph.errorBuffer) * 2)
			errorPtr = getCucharPtrFromByteSlicePtr(&ph.errorBuffer)
			errorBufferLen = cap(ph.errorBuffer)
		}
	}
	// rebuild error
	if errorBufferLen != 0 {
		err = errors.New(string(ph.errorBuffer[0:errorBufferLen]))
	}
	// rebuild result
	if resultBufferLen != 0 {
		decodeErr := ph.coder.Decode(bytes.NewBuffer(ph.resultBuffer[0:resultBufferLen]), result)
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

func (ph *PluginHandle)callEpELLoop(callPtr unsafe.Pointer, eventParam interface{}) (string, interface{}, error, bool) {
	// call entry point
	var eventId uint64 = 0
	var eventParamBufferLen int = cap(ph.eventParamBuffer) 
	var eventNameBufferLen int = cap(ph.eventNameBuffer)
	var errorBufferLen int = cap(ph.errorBuffer)
	eventParamPtr := getCucharPtrFromByteSlicePtr(&ph.eventParamBuffer)
	eventNamePtr := getCucharPtrFromByteSlicePtr(&ph.eventNameBuffer)
	errorPtr := getCucharPtrFromByteSlicePtr(&ph.errorBuffer)
	for {
		code := C.bridge_elloop_func(callPtr, C.ulonglong(ph.instanceId), getCulonglongPtrFromUint64Ptr(&eventId),
		    eventNamePtr, getCintPtrFromIntPtr(&eventNameBufferLen),
		    eventParamPtr, getCintPtrFromIntPtr(&eventParamBufferLen),
		    errorPtr, getCintPtrFromIntPtr(&errorBufferLen))
		if code == 0 {
			// ok
			break
		} else if code == 3 {
			// more event name buffer
			// glow buffer and recall
			ph.eventNameBuffer = make([]byte, 0, cap(ph.eventNameBuffer) * 2)
			eventNamePtr = getCucharPtrFromByteSlicePtr(&ph.eventNameBuffer)
			eventNameBufferLen = cap(ph.eventNameBuffer)
		} else if code == 4 {
			// more event param buffer
			// glow buffer and recall
			ph.eventParamBuffer = make([]byte, 0, cap(ph.eventParamBuffer) * 2)
			eventParamPtr = getCucharPtrFromByteSlicePtr(&ph.eventParamBuffer)
			eventParamBufferLen = cap(ph.eventParamBuffer)
		} else if code == 2 {
			// more error buffer
			// glow buffer and recall
			ph.errorBuffer = make([]byte, 0, cap(ph.errorBuffer) * 2)
			errorPtr = getCucharPtrFromByteSlicePtr(&ph.errorBuffer)
			errorBufferLen = cap(ph.errorBuffer)
		} else if code == 5 {
			// finish
			return "", eventParam, nil, true
		}
	}
	// rebuild error
	var err error = nil
	if errorBufferLen != 0 {
		err = errors.New(string(ph.errorBuffer[0:errorBufferLen]))
	}
	// rebuild event name
	var eventName string
	if eventNameBufferLen != 0 {
		eventName = string(ph.eventNameBuffer[0:eventNameBufferLen])
	}
	// rebuild event param
	if eventParamBufferLen != 0 {
		decodeErr := ph.coder.Decode(bytes.NewBuffer(ph.eventParamBuffer[0:eventParamBufferLen]), eventParam)
		if decodeErr != nil {
			if err == nil {
				err = decodeErr
			} else {
				err = errors.Wrap(err, decodeErr.Error())
			}
		}
	}
	return eventName, eventParam, err, false
}

type Plugger struct {
	loaded bool
	dynLoadLibs map[string]*dynLoadLib
	pluginHandles map[uint64]*PluginHandle
	idGen *generator.IdGenerator
}

func (p *Plugger) Load(pluginDirPath string) error {
	if p.loaded {
		return errors.New("plugins are already loaded")
	}
	pluginFiles := make([]string, 0, 0)
	if err := gatherPluginFiles(&pluginFiles, pluginDirPath, ".so"); err != nil {
		return err
	}
	for _, pf := range pluginFiles {
		dlHandle, err := dlOpen(pf, Now|Local)
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		isMatchName, err := dlHandle.dlSymbol("IsMatchName")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		newPlugin, err := dlHandle.dlSymbol("NewPlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		initPlugin, err := dlHandle.dlSymbol("InitPlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		startPlugin, err := dlHandle.dlSymbol("StartPlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		stopPlugin, err := dlHandle.dlSymbol("StopPlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		reloadPlugin, err := dlHandle.dlSymbol("ReloadPlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		finiPlugin, err := dlHandle.dlSymbol("FiniPlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		command, err := dlHandle.dlSymbol("Command")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		freePlugin, err := dlHandle.dlSymbol("FreePlugin")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		eventListenrLoop, err := dlHandle.dlSymbol("EventListenerLoop")
		if err != nil {
			errors.Fprint(os.Stdout, err)
			continue
		}
		dlLib := &dynLoadLib {
			dlHandle          : dlHandle,
			isMatchName       : isMatchName,
			newPlugin         : newPlugin,
			initPlugin        : initPlugin,
			startPlugin       : startPlugin,
			stopPlugin        : stopPlugin,
			reloadPlugin      : reloadPlugin,
			finiPlugin        : finiPlugin,
			command           : command,
			freePlugin        : freePlugin,
			eventListenerLoop : eventListenrLoop,
		}
		p.dynLoadLibs[pf] = dlLib
	}
	p.loaded = true
	return nil
}

func (p *Plugger) ExistsPlugins(pluginNames []string) []string {
	exists := make([]string, 0, 0)
	for _, pluginName := range pluginNames {
		if dlLib:= p.getDlLib(pluginName); dlLib == nil {
			continue
		}
		exists = append(exists, pluginName)
	}
	return exists
}

func (p *Plugger) NewPlugin(pluginName string, newResultFunc func() interface{}, newEventParamFunc func() interface{}) (*PluginHandle, error) {
	dlLib := p.getDlLib(pluginName)
	if dlLib == nil {
		return nil, errors.Errorf("not found plugin name (%v)", pluginName)
	}
	instanceId := p.idGen.Get()
	C.bridge_id_only(dlLib.newPlugin, C.ulonglong(instanceId))
	ph := &PluginHandle {
		instanceId: instanceId,
		dlLib: dlLib,
		pluginName: pluginName,
		errorBuffer: make([]byte, 0, definer.ErrorBufferSize),
		resultBuffer: make([]byte, 0, definer.ResultBufferSize),
		eventParamBuffer: make([]byte, 0, definer.EventParamBufferSize),
		eventNameBuffer: make([]byte, 0, definer.EventNameBufferSize),
		coder: coder.NewCoder(),
		newResultFunc : newResultFunc,
		newEventParamFunc : newEventParamFunc,
		eventHandlers : make(map[string]func(pluginHandle *PluginHandle, eventName string, eventParam interface{}, err error)),
	}
	go ph.eventListenerLoop()
	p.pluginHandles[instanceId] =  ph
	return ph, nil
}

func (p *Plugger) FreePlugin(pluginHandle *PluginHandle) error {
	dlLib := p.getDlLib(pluginHandle.pluginName)
	if dlLib == nil {
		return errors.Errorf("not found plugin name (%v)", pluginHandle.pluginName)
	}
	if _, ok := p.pluginHandles[pluginHandle.instanceId]; !ok {
		return errors.Errorf("not found plugin handle (%v)", pluginHandle.instanceId)
	}
	C.bridge_id_only(dlLib.freePlugin, C.ulonglong(pluginHandle.instanceId))
	delete(p.pluginHandles, pluginHandle.instanceId)
	return nil
}

func (p *Plugger) Free() {
	for instanceId, pluginHandle := range p.pluginHandles {
		C.bridge_id_only(pluginHandle.dlLib.freePlugin, C.ulonglong(instanceId))
	}
	p.pluginHandles = make(map[uint64]*PluginHandle)
	for _, dlLib := range p.dynLoadLibs {
		dlLib.dlHandle.dlClose()
	}
	p.dynLoadLibs = make(map[string]*dynLoadLib)
	p.loaded = false
}

func (p *Plugger) getDlLib(pluginName string) *dynLoadLib {
	for _, dlLib := range p.dynLoadLibs {
		if dlLib.pluginName != "" {
			if dlLib.pluginName == pluginName {
				return dlLib
			}
		} else {
			cspn := C.CString(pluginName)
			defer C.free(unsafe.Pointer(cspn))
			result := C.bridge_ismatch(dlLib.isMatchName, cspn)
			if result == 1 {
				dlLib.pluginName = pluginName
				return dlLib
			}
		}
	}
	return nil
}

func NewPlugger() *Plugger {
	p := new(Plugger)
	p.dynLoadLibs = make(map[string]*dynLoadLib)
	p.pluginHandles = make(map[uint64]*PluginHandle)
	p.idGen = generator.NewIdGenerator()
	return p
}

func getCucharPtrFromByteSlicePtr(value *[]byte) *C.char {
        return (*C.char)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(value)).Data))
}

func getCulonglongPtrFromUint64Ptr(value *uint64) *C.ulonglong {
        return (*C.ulonglong)(unsafe.Pointer(value))
}

func getCintPtrFromIntPtr(value *int) *C.int {
        return (*C.int)(unsafe.Pointer(value))
}

func ptrAndLenToBytes(ptr *C.char, ptrLen C.int) []byte {
        return C.GoBytes(unsafe.Pointer(ptr), ptrLen)
}
