package fabriceventmgr

import (
	"sync"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/eventmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type FabricEventManager struct {
	eventmgr.EventManagerBase
	ctx     *chaincodectx.FabricChaincodeCtx
	mapLock sync.RWMutex
}

func NewFabricEventManager(ctx *chaincodectx.FabricChaincodeCtx) *FabricEventManager {
	return &FabricEventManager{
		EventManagerBase: eventmgr.EventManagerBase{
			QuitChanMap: make(map[eventmgr.IEventRegistration]chan struct{}),
		},
		ctx:     ctx,
		mapLock: sync.RWMutex{},
	}
}

func (m *FabricEventManager) RegisterEvent(eventID string) (eventmgr.IEventRegistration, <-chan eventmgr.IEvent, error) {
	// Use Fabric's native channel client to register a chaincode event. The chaincode is specified in the context.
	rawReg, rawNotifier, err := m.ctx.ChannelClient.RegisterChaincodeEvent(m.ctx.ChaincodeID, eventID)
	if err != nil {
		return nil, nil, err
	}

	// Encapsulate the registration in an `eventmgr.IEventRegistration` object.
	fabricReg := &FabricEventRegistration{
		reg:     rawReg,
		eventID: eventID,
	}

	notifier := make(chan eventmgr.IEvent)
	quitChan := make(chan struct{})
	// Background task: wrap received Fabric events to `eventmgr.IEvent` objects.
	go func() {
		for {
			select {
			case event := <-rawNotifier:
				fabricEvent := (*FabricEvent)(event)
				notifier <- fabricEvent
			case <-quitChan:
				close(notifier)
				return
			}
		}
	}()

	// Record the registration. The quit handles are useful to stop the listening processes in the background.
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	m.QuitChanMap[fabricReg] = quitChan

	return fabricReg, notifier, nil
}

func (m *FabricEventManager) UnregisterEvent(reg eventmgr.IEventRegistration) error {
	// Cannot unregister non-Fabric registration with a Fabric event manager
	fabricReg := reg.(*FabricEventRegistration)

	// Unregister the event using the native channel client first to stop receiving events
	m.ctx.ChannelClient.UnregisterChaincodeEvent(fabricReg.reg)

	// Use the quit chan to stop the corresponding background process for event conversion
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	quitChan := m.QuitChanMap[fabricReg]
	quitChan <- struct{}{}

	// Now the quit chan entry is not useful. Close the quit chan and remove it from the map.
	close(quitChan)
	delete(m.QuitChanMap, fabricReg)

	return nil
}

type FabricEventRegistration struct {
	reg     fab.Registration
	eventID string
}

func (r *FabricEventRegistration) GetEventID() string {
	return r.eventID
}

type FabricEvent fab.CCEvent

func (e *FabricEvent) GetEventName() string {
	return e.EventName
}

func (e *FabricEvent) GetPayload() []byte {
	return e.Payload
}

func (e *FabricEvent) GetBlockNumber() uint64 {
	return e.BlockNumber
}

func (e *FabricEvent) GetTxID() string {
	return e.TxID
}
