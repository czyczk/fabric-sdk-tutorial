package eventmgr

type IEventManager interface {
	// RegisterEvent registers an event using the blockchain's built-in utilities (e.g. a channel client in Fabric). Events are emitted from a chaincode environment that's usually specified by the context of the event manager.
	//
	// Returns:
	//   the registration (used to unregister the event)
	//   the event channel
	RegisterEvent(eventID string) (IEventRegistration, <-chan IEvent, error)

	// UnregisterEvent unregisters a registration. The registration must be produced by the same event manager instance.
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
