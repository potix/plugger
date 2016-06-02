package common

type ReturnValue int

const (
	RVSuccess                  ReturnValue = 0
	RVNotImplemented                       = 1
	RVNeedGlowPluginNameBuffer             = 2
	RVNeedGlowResultBuffer                 = 3
	RVNeedGlowErrorBuffer                  = 4 
	RVNeedGlowEventNameBuffer              = 5
	RVNeedGlowEventParamBuffer             = 6
	RVFinishEventListenerLoop              = 7
)

type EventHandling int

const (
	EVHandling       EventHandling = 0
	EVNotImplemented               = 1
	EVNoHandling                   = 2
	EVEncodeError                  = 3
	EVDecodeError                  = 4
)
