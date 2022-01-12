package eventmgr

type IEventManager interface {
	RegisterEvent(eventID string) (IEventRegistration, <-chan IEvent, error)
	UnregisterEvent(reg IEventRegistration) error
}

type EventManagerBase struct {
	QuitChanMap map[IEventRegistration]chan struct{}
}

type IEventRegistration interface {
	GetEventID() string
}

type IEvent interface {
	GetEventName() string
	GetPayload() []byte
	GetBlockNumber() uint64
	GetTxID() string
}
